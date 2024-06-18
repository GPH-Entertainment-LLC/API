package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"
	"xo-packs/core"
	"xo-packs/db"
	"xo-packs/model"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type ItemRepository interface {
	CreateItem(context.Context, *model.Item) (*model.Item, error)
	GetItem(context.Context, uint64) (*model.Item, error)
	GetItems(context.Context, []uint64, string) ([]model.Item, error)
	UserOwnsItem(context.Context, *string, uint64) (bool, error)
	AddItemCategories(context.Context, []*model.ItemCategory) error
	AddUserItems(context.Context, string, []uint64) error
	PatchItem(context.Context, uint64, map[string]interface{}, string) (*model.Item, error)
	DeleteUserItems(context.Context, []uint64, string) error
	DeleteItems(context.Context, []uint64, string) error
	ClearItemCategoryCache(context.Context) error
	ClearUserItemCache(context.Context, string) error
	ClearVendorItemCache(context.Context, string) error
	ClearItemCache(context.Context, []string) error
	GetItemSignCreds(context.Context) (string, string, error)
}

type ItemRepoImpl struct {
	db    *sqlx.DB
	cache *redis.Client
}

func NewItemRepo(db *sqlx.DB, cache *redis.Client) ItemRepository {
	return &ItemRepoImpl{db: db, cache: cache}
}

type ItemError struct {
	message string
}

func (e *ItemError) Error() string {
	return e.message
}

func (r *ItemRepoImpl) CreateItem(c context.Context, item *model.Item) (*model.Item, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			if err = tx.Rollback(); err != nil {
				fmt.Println(err)
			}
		}
	}()

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Insert(db.SCHEMA_ITEMS).
		Columns(core.ModelColumns(item)...).
		Values(core.StructValues(item)...).
		Suffix("RETURNING \"id\"").
		ToSql()
	if err != nil {
		return nil, err
	}

	insertedId := new(uint64)
	err = tx.QueryRowContext(ctx, query, args...).Scan(insertedId)
	if err != nil {
		return nil, err
	}
	if insertedId == nil {
		return nil, &core.ErrorResp{Message: "critical error; unable to add item"}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	item, err = r.GetItem(c, *insertedId)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func (r *ItemRepoImpl) GetItem(c context.Context, itemId uint64) (*model.Item, error) {
	val, err := r.cache.Get(c, db.KEY_ITEM+(fmt.Sprintf("%v", itemId))).Result()
	if err != nil {
		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()

		psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
		query, args, err := psql.
			Select("*").
			From(db.SCHEMA_ITEMS).
			Where(squirrel.Eq{"id": itemId, "deleted_at": nil}).
			ToSql()
		if err != nil {
			return nil, err
		}

		rows, err := r.db.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}

		item := model.Item{}
		defer rows.Close()
		for rows.Next() {
			if err := rows.StructScan(&item); err != nil {
				return nil, err
			}
		}
		if item.ID != nil {
			itemBytes, err := json.Marshal(item)
			if err != nil {
				return nil, err
			}
			if err = r.cache.Set(c, db.KEY_ITEM+(fmt.Sprintf("%v", itemId)), itemBytes, time.Duration(3600)*time.Second).Err(); err != nil {
				return nil, err
			}
		}
		return &item, nil
	} else {
		item := model.Item{}
		if err = json.Unmarshal([]byte(val), &item); err != nil {
			return nil, err
		}
		return &item, nil
	}
}

func (r *ItemRepoImpl) GetItems(c context.Context, itemIds []uint64, vendorId string) ([]model.Item, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Select("*").
		From(db.SCHEMA_ITEMS).
		Where(squirrel.Eq{"id": itemIds, "vendor_id": vendorId, "deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	items := []model.Item{}
	item := model.Item{}
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *ItemRepoImpl) AddItemCategories(c context.Context, categories []*model.ItemCategory) error {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query := psql.
		Insert(db.SCHEMA_ITEM_CATEGORIES).
		Columns("item_id", "category_id")

	for _, category := range categories {
		query = query.Values(core.StructValues(category)...)
	}
	queryStr, args, err := query.ToSql()
	if err != nil {
		return nil
	}
	_, err = tx.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *ItemRepoImpl) UserOwnsItem(c context.Context, uid *string, itemId uint64) (bool, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return false, err
	}

	defer func() {
		tx.Commit()
	}()

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Select("id").
		From(db.SCHEMA_USER_ITEMS).
		Where(squirrel.Eq{"uid": *uid, "item_id": itemId}).
		ToSql()
	if err != nil {
		return false, err
	}
	rows, err := tx.QueryxContext(ctx, query, args...)
	if err != nil {
		return false, err
	}
	exists := false
	id := uint64(0)
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return false, err
		}
		exists = true
	}
	return exists, nil
}

func (r *ItemRepoImpl) AddUserItems(c context.Context, uid string, itemIds []uint64) error {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	if len(itemIds) <= 0 {
		return nil
	}

	currentTime := time.Now()
	now := currentTime.Format("2006-01-02 15:04:05")
	userItems := make([]model.UserItem, len(itemIds))
	for i, id := range itemIds {
		userItemId := id
		userItem := model.UserItem{
			Uid:        &uid,
			ItemId:     &userItemId,
			AcquiredAt: &now,
		}
		userItems[i] = userItem
	}

	if len(userItems) != len(itemIds) {
		return &core.ErrorResp{
			Message: "critical error: unable to create user items corresponding to item ids",
		}
	}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query := psql.
		Insert(db.SCHEMA_USER_ITEMS).
		Columns(core.ModelColumns(userItems[0])...)

	for _, userItem := range userItems {
		query = query.Values(core.StructValues(userItem)...)
	}

	queryStr, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *ItemRepoImpl) PatchItem(c context.Context, itemId uint64, itemPatchMap map[string]interface{}, authorizedUid string) (*model.Item, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			if err = tx.Rollback(); err != nil {
				fmt.Println(err)
			}
		}
	}()

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Update(db.SCHEMA_ITEMS).
		SetMap(itemPatchMap).
		Where(squirrel.Eq{"id": itemId, "deleted_at": nil, "vendor_id": authorizedUid}).
		ToSql()
	if err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected > 0 {
		err = tx.Commit()
		if err != nil {
			return nil, err
		}

		// remove cached item
		if err = r.cache.Del(c, db.KEY_ITEM+fmt.Sprintf("%v", itemId)).Err(); err != nil {
			return nil, err
		}

		// getting new updated item
		item, err := r.GetItem(c, itemId)
		if err != nil {
			return nil, err
		}
		return item, nil
	} else {
		err = &ItemError{"An error occurred during the update. Item either is not active or could not be mapped using the given patch."}
		return nil, err
	}
}

func (r *ItemRepoImpl) DeleteUserItems(c context.Context, userItemIds []uint64, uid string) error {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	now := time.Now().Format("2006-01-02 15:04:05")
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Update(db.SCHEMA_USER_ITEMS).
		Set("removed_at", now).
		Where(squirrel.Eq{"id": userItemIds, "uid": uid}).
		ToSql()
	if err != nil {
		return err
	}
	rows, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := rows.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected > 0 {
		return nil
	} else {
		return &core.ErrorResp{Message: "User items could not be deleted or do not exist"}
	}
}

func (r *ItemRepoImpl) DeleteItems(c context.Context, itemIds []uint64, vendorId string) error {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	now := time.Now().Format("2006-01-02 15:04:05")
	query, args, err := psql.
		Update(db.SCHEMA_ITEMS).
		SetMap(map[string]interface{}{"active": false, "deleted_at": now}).
		Where(squirrel.Eq{"id": itemIds, "vendor_id": vendorId}).
		ToSql()
	if err != nil {
		return err
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected > 0 {
		return nil
	} else {
		err = &ItemError{
			fmt.Sprintf("An error occurred while deleting items. Item either does not exist for this given vendor or cannot be deleted. \n%v", err.Error()),
		}
		return err
	}
}

func (r *ItemRepoImpl) ClearItemCategoryCache(c context.Context) error {
	return r.cache.Del(c, db.KEY_CATEGORY+"item_categories").Err()
}

func (r *ItemRepoImpl) ClearUserItemCache(c context.Context, uid string) error {
	keysToDelete := []string{
		db.KEY_USER_ITEMS + uid,
		db.KEY_USER_ITEM_AMOUNT + uid,
		fmt.Sprintf("/user/items/%v*", uid),
	}

	allKeys := []string{}
	for _, key := range keysToDelete {
		keys, err := r.cache.Keys(c, key).Result()
		if err != nil {
			return err
		}
		allKeys = append(allKeys, keys...)
	}

	for _, key := range allKeys {
		if err := r.cache.Del(c, key).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (r *ItemRepoImpl) ClearVendorItemCache(c context.Context, vendorId string) error {
	keys, err := r.cache.Keys(c, fmt.Sprintf("/vendor/items/%v*", vendorId)).Result()
	if err != nil {
		return err
	}
	for _, key := range keys {
		if err = r.cache.Del(c, key).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (r *ItemRepoImpl) ClearItemCache(c context.Context, itemKeys []string) error {
	return r.cache.Del(c, itemKeys...).Err()
}

func (r *ItemRepoImpl) GetItemSignCreds(c context.Context) (string, string, error) {
	var pemKey string
	var keyPairID string

	pemKey, err := r.cache.Get(c, db.KEY_PEM).Result()
	if err != nil {
		pemKey = os.Getenv("PRIVATE_KEY")
		pemKey, err = core.GetSecret(pemKey)
		if err != nil {
			return "", "", err
		}

		if err = r.cache.Set(c, db.KEY_PEM, pemKey, time.Duration(3600*time.Second)).Err(); err != nil {
			return "", "", err
		}
	}

	keyPairID, err = r.cache.Get(c, db.KEY_KEY_PAIR_ID).Result()
	if err != nil {
		keyPairID = os.Getenv("KEY_PAIR_ID")
		keyPairID, err = core.GetSecret(keyPairID)
		if err != nil {
			return "", "", err
		}

		if err = r.cache.Set(c, db.KEY_KEY_PAIR_ID, keyPairID, time.Duration(3600*time.Second)).Err(); err != nil {
			return "", "", err
		}
	}

	return pemKey, keyPairID, nil
}
