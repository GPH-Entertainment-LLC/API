package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"xo-packs/core"
	"xo-packs/db"
	"xo-packs/model"
	"xo-packs/query"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type UserRepository interface {
	CreateUser(context.Context, *model.User) (*model.User, error)
	GetUser(context.Context, string, bool) (*model.User, error)
	GetUserItem(context.Context, uint64) (*model.UserItem, error)
	WithdrawalUserItem(context.Context, uint64) (*uint64, error)
	GetUserItemWithdrawal(context.Context, uint64) (*model.ItemWithdrawal, error)
	GetUserPackPage(context.Context, string, uint64, string, string, string, string) ([]*model.PageUserPack, *uint64, error)
	GetUserItemPage(context.Context, string, uint64, string, string, string, string) ([]*model.PageUserItem, *uint64, error)
	GetUserFavoritesPage(*gin.Context, string, uint64, string, string) ([]*model.PageUserFavorite, *uint64, error)
	GetUserByUsername(context.Context, string) (*model.User, error)
	GetUserPackCategories(context.Context) ([]string, error)
	GetUserPackSortMappings(context.Context) (map[string]string, error)
	GetUserItemCategories(context.Context) ([]string, error)
	GetUserItemSortMappings(context.Context) (map[string]string, error)
	PatchUser(context.Context, string, string, map[string]interface{}) error
	DeleteUser(context.Context, string, string) error
	ClearUserCache(context.Context, string, string) error
	FlushCache(context.Context) error
	ClearFavoriteCache(context.Context, string, string) error
	GetFavorite(context.Context, string, string) (*model.Favorite, error)
	AddFavorite(context.Context, string, string) (*model.Favorite, error)
	RemoveFavorite(context.Context, string, string) error
}

type UserRepoImpl struct {
	db    *sqlx.DB
	cache *redis.Client
}

type UserError struct {
	message string
}

func (e *UserError) Error() string {
	return e.message
}

func NewUserRepo(db *sqlx.DB, cache *redis.Client) UserRepository {
	return &UserRepoImpl{db: db, cache: cache}
}

func (r *UserRepoImpl) CreateUser(c context.Context, newUser *model.User) (*model.User, error) {
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

	if newUser.Username != nil {
		user, err := r.GetUserByUsername(c, *newUser.Username)
		if err != nil {
			return nil, err
		}
		if user.Uid != nil {
			return nil, &UserError{message: "Error, username already taken"}
		}
	}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Insert(db.SCHEMA_USERS).
		Columns(core.ModelColumns(newUser)...).
		Values(core.StructValues(newUser)...).
		ToSql()
	if err != nil {
		return newUser, err
	}

	// Execute the query
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func (r *UserRepoImpl) GetUser(c context.Context, uid string, private bool) (*model.User, error) {
	fieldList := core.UserFieldList
	cacheKey := db.KEY_USER + uid
	if private {
		fieldList = core.PrivateUserFieldList
		cacheKey = db.KEY_PRIVATE_USER + uid
	}
	val, err := r.cache.Get(c, cacheKey).Result()
	if err != nil {
		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()

		tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return nil, err
		}

		defer func() {
			if err = tx.Commit(); err != nil {
				fmt.Println(err)
			}
		}()

		psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
		query, args, err := psql.
			Select(fieldList...).
			From(db.SCHEMA_USERS).
			Where(squirrel.Eq{"uid": uid, "active": true}).
			ToSql()
		if err != nil {
			return nil, err
		}

		rows, err := tx.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}

		user := model.User{}
		defer rows.Close()
		for rows.Next() {
			if err := rows.StructScan(&user); err != nil {
				return nil, err
			}
		}

		if user.Uid != nil {
			userBytes, err := json.Marshal(user)
			if err != nil {
				return nil, err
			}

			if err = r.cache.Set(c, cacheKey, userBytes, time.Duration(3600)*time.Second).Err(); err != nil {
				return nil, err
			}
		}
		return &user, nil
	} else {
		user := model.User{}
		if err = json.Unmarshal([]byte(val), &user); err != nil {
			return nil, err
		}
		return &user, nil
	}
}

func (r *UserRepoImpl) GetUserItem(c context.Context, userItemId uint64) (*model.UserItem, error) {
	val, err := r.cache.Get(c, db.KEY_USER_ITEM+fmt.Sprintf("%v", userItemId)).Result()
	if err != nil {
		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()

		tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return nil, err
		}

		defer func() {
			if err = tx.Commit(); err != nil {
				fmt.Println(err)
			}
		}()

		userItem := model.UserItem{}
		psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
		query, args, err := psql.
			Select("id", "uid", "item_id", "removed_at", "acquired_at", "expired_at").
			From(db.SCHEMA_USER_ITEMS).
			Where(squirrel.Eq{"id": userItemId}).
			ToSql()
		if err != nil {
			return nil, err
		}

		rows, err := tx.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}

		defer rows.Close()
		for rows.Next() {
			if err = rows.StructScan(&userItem); err != nil {
				return nil, err
			}
		}

		userItemBytes, err := json.Marshal(userItem)
		if err != nil {
			return nil, err
		}

		if err = r.cache.Set(ctx, db.KEY_USER_ITEM+fmt.Sprintf("%v", userItemId), userItemBytes, time.Duration(3600*time.Second)).Err(); err != nil {
			return nil, err
		}
		return &userItem, nil
	} else {
		userItem := model.UserItem{}
		if err = json.Unmarshal([]byte(val), &userItem); err != nil {
			return nil, err
		}
		return &userItem, nil
	}
}

func (r *UserRepoImpl) WithdrawalUserItem(c context.Context, userItemId uint64) (*uint64, error) {
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

	now := time.Now().Format("2006-01-02 15:04:05")
	itemWithdrawal := model.ItemWithdrawal{
		UserItemId:  &userItemId,
		WithdrawnAt: &now,
	}
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Insert(db.SCHEMA_ITEM_WITHDRAWALS).
		Columns(core.ModelColumns(itemWithdrawal)...).
		Values(core.StructValues(itemWithdrawal)...).
		Suffix("RETURNING ID").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	withdrawalId := new(uint64)
	for rows.Next() {
		if err = rows.Scan(withdrawalId); err != nil {
			return nil, err
		}
	}
	if withdrawalId == nil || (withdrawalId != nil && *withdrawalId <= 0) {
		return nil, &core.ErrorResp{
			Message: "unable to withdrawal item at this time",
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return withdrawalId, nil
}

func (r *UserRepoImpl) GetUserItemWithdrawal(c context.Context, withdrawalId uint64) (*model.ItemWithdrawal, error) {
	val, err := r.cache.Get(c, db.KEY_ITEM_WITHDRAWALS+fmt.Sprintf("%v", withdrawalId)).Result()
	if err != nil {
		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()

		tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return nil, err
		}

		defer func() {
			if err = tx.Commit(); err != nil {
				fmt.Println(err)
			}
		}()

		itemWithdrawal := model.ItemWithdrawal{}
		psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
		query, args, err := psql.
			Select("id", "user_item_id", "withdrawn_at", "fulfilled_at").
			From(db.SCHEMA_ITEM_WITHDRAWALS).
			Where(squirrel.Eq{"id": withdrawalId}).
			ToSql()
		if err != nil {
			return nil, err
		}

		rows, err := tx.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			if err = rows.StructScan(&itemWithdrawal); err != nil {
				return nil, err
			}
		}

		itemWithdrawalBytes, err := json.Marshal(itemWithdrawal)
		if err != nil {
			return nil, err
		}

		if err = r.cache.Set(ctx, db.KEY_ITEM_WITHDRAWALS+fmt.Sprintf("%v", withdrawalId), itemWithdrawalBytes, time.Duration(3600*time.Second)).Err(); err != nil {
			return nil, err
		}

		return &itemWithdrawal, nil
	} else {
		itemWithdrawal := model.ItemWithdrawal{}
		if err = json.Unmarshal([]byte(val), &itemWithdrawal); err != nil {
			return nil, err
		}
		return &itemWithdrawal, nil
	}
}

func (r *UserRepoImpl) GetUserPackPage(
	c context.Context, uid string, pageNumber uint64, sortBy string, filterOn string, searchStr string, urlPath string) ([]*model.PageUserPack, *uint64, error) {

	val, err := r.cache.Get(c, urlPath).Result()
	if err != nil {
		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()

		tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return nil, nil, err
		}

		defer func() {
			if err = tx.Commit(); err != nil {
				fmt.Println(err)
			}
		}()

		pageSizeStr := os.Getenv("USER_PACKS_PAGE_SIZE")
		pageSize, err := strconv.ParseUint(pageSizeStr, 10, 64)
		if err != nil {
			return nil, nil, err
		}

		var rows *sqlx.Rows
		if sortBy != "" {
			if filterOn != "" {
				rows, err = tx.QueryxContext(ctx, fmt.Sprintf(
					query.UserPackPageSortByFilterOn,
					uid,
					filterOn,
					searchStr,
					searchStr,
					searchStr,
					searchStr,
					searchStr,
					sortBy,
					pageSize,
					(pageSize)*(pageNumber)),
				)
			} else {
				rows, err = tx.QueryxContext(ctx, fmt.Sprintf(
					query.UserPackPageSortBy,
					uid,
					searchStr,
					searchStr,
					searchStr,
					searchStr,
					searchStr,
					sortBy,
					pageSize,
					(pageSize)*(pageNumber)),
				)
			}
		} else if filterOn != "" {
			rows, err = tx.QueryxContext(ctx, fmt.Sprintf(
				query.UserPackPageFilterOn,
				uid,
				searchStr,
				searchStr,
				searchStr,
				searchStr,
				searchStr,
				filterOn,
				pageSize,
				(pageSize)*(pageNumber)),
			)
		} else {
			rows, err = tx.QueryxContext(ctx, fmt.Sprintf(
				query.UserPackPageDefault,
				uid,
				searchStr,
				searchStr,
				searchStr,
				searchStr,
				searchStr,
				pageSize,
				(pageSize)*(pageNumber)),
			)
		}
		if err != nil {
			return nil, nil, err
		}

		defer rows.Close()
		pagePacks := []*model.PageUserPack{}
		for rows.Next() {
			pagePack := model.PageUserPack{}
			if err := rows.StructScan(&pagePack); err != nil {
				return nil, nil, err
			}
			if pagePack.RawCategories != nil {
				categories := strings.Split(*pagePack.RawCategories, ",")
				pagePack.Categories = categories
			}
			if pagePack.RawPackIds != nil {
				rawPackIds := strings.Split(*pagePack.RawPackIds, ",")
				fmt.Println(rawPackIds)
				packIds := make([]uint64, len(rawPackIds))
				for i, v := range rawPackIds {
					packId, err := strconv.ParseUint(v, 10, 64)
					if err != nil {
						return nil, nil, err
					}
					packIds[i] = packId
				}
				pagePack.PackIds = packIds
			}
			pagePacks = append(pagePacks, &pagePack)
		}
		packAmount := new(uint64)
		if len(pagePacks) <= 0 {
			*packAmount = 0
		} else {
			packAmount = pagePacks[len(pagePacks)-1].TotalPacksAmount
		}

		pagePackBytes, err := json.Marshal(pagePacks)
		if err != nil {
			return nil, nil, err
		}
		if err = r.cache.Set(c, urlPath, pagePackBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, nil, err
		}

		return pagePacks, packAmount, nil
	} else {
		pagePacks := []*model.PageUserPack{}
		if err = json.Unmarshal([]byte(val), &pagePacks); err != nil {
			return nil, nil, err
		}
		packAmount := new(uint64)
		if len(pagePacks) <= 0 {
			*packAmount = 0
		} else {
			packAmount = pagePacks[len(pagePacks)-1].TotalPacksAmount
		}
		return pagePacks, packAmount, nil
	}
}

func (r *UserRepoImpl) GetUserItemPage(
	c context.Context, uid string, pageNumber uint64, sortBy string, filterOn string, searchStr string, urlPath string) ([]*model.PageUserItem, *uint64, error) {

	val, err := r.cache.Get(c, urlPath).Result()
	if err != nil {
		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()

		tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return nil, nil, err
		}

		defer func() {
			if err = tx.Commit(); err != nil {
				fmt.Println(err)
			}
		}()

		pageSizeStr := os.Getenv("USER_ITEMS_PAGE_SIZE")
		pageSize, err := strconv.ParseUint(pageSizeStr, 10, 64)
		if err != nil {
			return nil, nil, err
		}

		var rows *sqlx.Rows
		if sortBy != "" {
			if filterOn != "" {
				rows, err = tx.QueryxContext(ctx, fmt.Sprintf(
					query.UserItemPageSortByFilterOn,
					uid,
					filterOn,
					searchStr,
					searchStr,
					searchStr,
					searchStr,
					searchStr,
					sortBy,
					pageSize,
					(pageSize)*(pageNumber)),
				)
			} else {
				fmt.Printf(query.UserItemPageSortBy, uid, sortBy, pageSize, (pageSize)*(pageNumber))
				rows, err = tx.QueryxContext(ctx, fmt.Sprintf(
					query.UserItemPageSortBy,
					uid,
					searchStr,
					searchStr,
					searchStr,
					searchStr,
					searchStr,
					sortBy,
					pageSize, (pageSize)*(pageNumber)),
				)
			}
		} else if filterOn != "" {
			rows, err = tx.QueryxContext(ctx, fmt.Sprintf(
				query.UserItemPageFilterOn,
				uid,
				filterOn,
				searchStr,
				searchStr,
				searchStr,
				searchStr,
				searchStr,
				pageSize,
				(pageSize)*(pageNumber)),
			)
		} else {
			rows, err = tx.QueryxContext(ctx, fmt.Sprintf(
				query.UserItemPageDefault,
				uid,
				searchStr,
				searchStr,
				searchStr,
				searchStr,
				searchStr,
				pageSize,
				(pageSize)*(pageNumber)),
			)
		}
		if err != nil {
			return nil, nil, err
		}

		defer rows.Close()
		pageItems := []*model.PageUserItem{}
		for rows.Next() {
			pageItem := model.PageUserItem{}
			if err := rows.StructScan(&pageItem); err != nil {
				return nil, nil, err
			}
			if pageItem.RawCategories != nil {
				categories := strings.Split(*pageItem.RawCategories, ",")
				pageItem.Categories = categories
			}
			pageItems = append(pageItems, &pageItem)
		}
		itemAmount := new(uint64)
		if len(pageItems) <= 0 {
			*itemAmount = 0
		} else {
			itemAmount = pageItems[len(pageItems)-1].TotalItemsAmount
		}

		pageItemBytes, err := json.Marshal(pageItems)
		if err != nil {
			return nil, nil, err
		}
		if err = r.cache.Set(c, urlPath, pageItemBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, nil, err
		}
		return pageItems, itemAmount, nil
	} else {
		pageItems := []*model.PageUserItem{}
		if err = json.Unmarshal([]byte(val), &pageItems); err != nil {
			return nil, nil, err
		}
		itemAmount := new(uint64)
		if len(pageItems) <= 0 {
			*itemAmount = 0
		} else {
			itemAmount = pageItems[len(pageItems)-1].TotalItemsAmount
		}
		return pageItems, itemAmount, nil
	}
}

func (r *UserRepoImpl) GetUserFavoritesPage(c *gin.Context, uid string, pageNumber uint64, searchStr string, urlPath string) ([]*model.PageUserFavorite, *uint64, error) {
	val, err := r.cache.Get(c, urlPath).Result()
	if err != nil {
		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()

		tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return nil, nil, err
		}

		defer func() {
			if err = tx.Commit(); err != nil {
				fmt.Println(err)
			}
		}()

		pageSizeStr := os.Getenv("USER_FAVORITES_PAGE_SIZE")
		pageSize, err := strconv.ParseUint(pageSizeStr, 10, 64)
		if err != nil {
			return nil, nil, err
		}

		rows, err := tx.QueryxContext(ctx, fmt.Sprintf(
			query.UserFavoritePageDefault,
			uid,
			searchStr,
			searchStr,
			searchStr,
			searchStr,
			pageSize,
			pageSize*(pageNumber)),
		)
		if err != nil {
			return nil, nil, err
		}

		defer rows.Close()
		pageFavorites := []*model.PageUserFavorite{}
		for rows.Next() {
			pageFavorite := model.PageUserFavorite{}
			if err = rows.StructScan(&pageFavorite); err != nil {
				return nil, nil, err
			}
			pageFavorites = append(pageFavorites, &pageFavorite)
		}
		favoriteAmount := new(uint64)
		if len(pageFavorites) <= 0 {
			*favoriteAmount = 0
		} else {
			favoriteAmount = pageFavorites[len(pageFavorites)-1].FavoritesAmount
		}

		favoriteBytes, err := json.Marshal(pageFavorites)
		if err != nil {
			return nil, nil, err
		}
		if err = r.cache.Set(c, urlPath, favoriteBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, nil, err
		}
		return pageFavorites, favoriteAmount, nil

	} else {
		pageFavorites := []*model.PageUserFavorite{}
		if err = json.Unmarshal([]byte(val), &pageFavorites); err != nil {
			return nil, nil, err
		}
		favoriteAmount := new(uint64)
		if len(pageFavorites) <= 0 {
			*favoriteAmount = 0
		} else {
			favoriteAmount = pageFavorites[len(pageFavorites)-1].FavoritesAmount
		}
		return pageFavorites, favoriteAmount, nil
	}
}

func (r *UserRepoImpl) GetUserPackCategories(c context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = tx.Commit(); err != nil {
			fmt.Println(err)
		}
	}()
	return core.GetCategories(ctx, "pack_categories", tx, err)
}

func (r *UserRepoImpl) GetUserPackSortMappings(c context.Context) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = tx.Commit(); err != nil {
			fmt.Println(err)
		}
	}()
	category := "user packs"
	return core.GetSortMappings(ctx, &category, tx, err)
}

func (r *UserRepoImpl) GetUserItemCategories(c context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}

	defer func() {
		tx.Commit()
	}()

	return core.GetCategories(ctx, "item_categories", tx, err)
}

func (r *UserRepoImpl) GetUserItemSortMappings(c context.Context) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}

	defer func() {
		tx.Commit()
	}()

	category := "user items"
	return core.GetSortMappings(ctx, &category, tx, err)
}

func (r *UserRepoImpl) GetUserByUsername(c context.Context, username string) (*model.User, error) {
	val, err := r.cache.Get(c, db.KEY_USERNAME+username).Result()
	if err != nil {
		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()

		tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return nil, err
		}

		defer func() {
			if err = tx.Commit(); err != nil {
				fmt.Println(err)
			}
		}()

		psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
		query, args, err := psql.
			Select(core.UserFieldList...).
			From(db.SCHEMA_USERS).
			Where(squirrel.Eq{"username": username, "active": true}).
			ToSql()
		if err != nil {
			return nil, err
		}

		rows, err := tx.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}

		defer rows.Close()
		user := model.User{}
		for rows.Next() {
			if err := rows.StructScan(&user); err != nil {
				return nil, err
			}
		}

		if user.Uid != nil {
			userBytes, err := json.Marshal(user)
			if err != nil {
				return nil, err
			}

			if err = r.cache.Set(c, db.KEY_USERNAME+username, userBytes, time.Duration(3600)*time.Second).Err(); err != nil {
				return nil, err
			}
		}

		return &user, nil
	} else {
		user := model.User{}
		if err = json.Unmarshal([]byte(val), &user); err != nil {
			return nil, err
		}
		return &user, nil
	}
}

func (r *UserRepoImpl) PatchUser(c context.Context, uid string, username string, userPatchMap map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
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
		Update(db.SCHEMA_USERS).
		SetMap(userPatchMap).
		Where(squirrel.Eq{"uid": uid}).
		ToSql()
	if err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected > 0 {
		err = tx.Commit()
		if err != nil {
			return err
		}
		return nil
	}
	err = &UserError{message: fmt.Sprintf("User with UID: %v does not exist", uid)}
	return err
}

func (r *UserRepoImpl) DeleteUser(c context.Context, uid string, username string) error {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if err = tx.Rollback(); err != nil {
				fmt.Println(err)
			}
		}
	}()

	now := time.Now().Format("2006-01-02 15:04:05")
	fieldMap := map[string]interface{}{"active": false, "deleted_at": now}
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Update(db.SCHEMA_USERS).
		SetMap(fieldMap).
		Where(squirrel.Eq{"uid": uid}).
		ToSql()
	if err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected > 0 {
		err = tx.Commit()
		if err != nil {
			return err
		}
		return nil
	}
	err = &UserError{message: fmt.Sprintf("User with UID: %v does not exist", uid)}
	return err
}

func (r *UserRepoImpl) FlushCache(c context.Context) error {
	return r.cache.FlushAll(c).Err()
}

func (r *UserRepoImpl) ClearUserCache(c context.Context, uid string, username string) error {
	keysToDelete := []string{
		db.KEY_USERNAME + username,
		db.KEY_PRIVATE_USER + uid,
		db.KEY_USER + uid,
		"/vendor/" + uid,
		"/vendors*",
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

func (r *UserRepoImpl) ClearFavoriteCache(c context.Context, uid string, vendorId string) error {
	keysToDelete := []string{
		fmt.Sprintf("/user/favorites/%v*", uid),
		(uid + db.KEY_FAVORITES + vendorId),
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

func (r *UserRepoImpl) GetFavorite(c context.Context, uid string, vendorId string) (*model.Favorite, error) {
	val, err := r.cache.Get(c, (uid + db.KEY_FAVORITES + vendorId)).Result()
	if err != nil {
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
			Select("id", "uid", "vendor_id", "favorited_at").
			From(db.SCHEMA_FAVORITES).
			Where(squirrel.Eq{"uid": uid, "vendor_id": vendorId}).
			ToSql()
		if err != nil {
			return nil, err
		}

		rows, err := tx.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}

		favorite := model.Favorite{}
		defer rows.Close()
		for rows.Next() {
			if err = rows.StructScan(&favorite); err != nil {
				return nil, err
			}
		}

		favoriteBytes, err := json.Marshal(favorite)
		if err != nil {
			return nil, err
		}
		if err = r.cache.Set(c, (uid + db.KEY_FAVORITES + vendorId), favoriteBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, err
		}
		return &favorite, nil
	} else {
		favorite := model.Favorite{}
		if err = json.Unmarshal([]byte(val), &favorite); err != nil {
			return nil, err
		}
		return &favorite, nil
	}
}

func (r *UserRepoImpl) AddFavorite(c context.Context, uid string, vendorId string) (*model.Favorite, error) {
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

	now := time.Now().Format("2006-01-02 15:04:05")
	favorite := model.Favorite{Uid: &uid, VendorId: &vendorId, FavoritedAt: &now}
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Insert(db.SCHEMA_FAVORITES).
		Columns(core.ModelColumns(favorite)...).
		Values(core.StructValues(favorite)...).
		ToSql()
	if err != nil {
		return nil, &core.ErrorResp{Message: "Error, can only favorite a valid vendor"}
	}

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected <= 0 {
		return nil, &core.ErrorResp{Message: "Error, unable to insert favorite record"}
	}

	// update vendor favorites amount
	query, args, err = psql.
		Update(db.SCHEMA_VENDORS).
		Set("favorites_amount", squirrel.Expr("favorites_amount + 1")).
		Where(squirrel.Eq{"uid": vendorId}).
		ToSql()
	if err != nil {
		return nil, err
	}

	result, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected < 0 {
		return nil, &core.ErrorResp{Message: "Error, favorites amount could not be updated. Vendor either is invalid or does not exist"}
	} else {
		if err = tx.Commit(); err != nil {
			return nil, err
		}
		return &favorite, nil
	}
}

func (r *UserRepoImpl) RemoveFavorite(c context.Context, uid string, vendorId string) error {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
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
		Delete(db.SCHEMA_FAVORITES).
		Where(squirrel.Eq{"uid": uid, "vendor_id": vendorId}).
		ToSql()
	if err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	// update vendor favorites amount
	query, args, err = psql.
		Update(db.SCHEMA_VENDORS).
		Set("favorites_amount", squirrel.Expr("case when favorites_amount > 0 then favorites_amount - 1 else 0 end")).
		Where(squirrel.Eq{"uid": vendorId}).
		ToSql()
	if err != nil {
		return err
	}

	result, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected < 0 {
		return &core.ErrorResp{Message: "Error, favorites amount could not be updated. Vendor either is invalid or does not exist"}
	} else {
		if err = tx.Commit(); err != nil {
			return err
		}
		return nil
	}
}
