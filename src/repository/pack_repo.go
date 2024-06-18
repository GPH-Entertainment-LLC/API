package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	"xo-packs/core"
	"xo-packs/db"
	"xo-packs/model"
	"xo-packs/query"

	"github.com/redis/go-redis/v9"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type PackRepository interface {
	CreatePackConfig(context.Context, *model.PackConfig) (*model.PackConfig, error)
	AddPackItemConfigs(context.Context, []*model.PackItemConfig) error
	BuyPacks(context.Context, string, *model.PackConfig, float64, *model.TokenCurrencyRate) (*model.PackBoughtResp, error)
	OpenPack(context.Context, uint64, string) (*model.Pack, error)
	GetPack(context.Context, uint64) (*model.Pack, error)
	GetUserPackAmount(context.Context, string) (*uint64, error)
	GetPackConfig(context.Context, uint64) (*model.PackConfig, error)
	GetPackItems(context.Context, uint64) ([]*model.PackItemConfigExpanded, error)
	AddPackCategories(context.Context, []*model.PackCategory) error
	GetPackItemConfigs(context.Context, uint64) ([]*model.PackItemConfig, error)
	GetActivePackItems(context.Context, []uint64) ([]string, []string, error)
	GetPacksContainingItems(context.Context, []uint64) ([]string, []string, error)
	GetPackItemsPreview(context.Context, uint64) ([]*model.PackItemPreview, error)
	UploadPacks(context.Context, []*model.PackFact, uint64) ([]uint64, error)
	UploadPackItems(context.Context, []*model.PackItemFact) error
	UpdateVendorPackAmount(context.Context, string, int, *sqlx.Tx) error
	PatchPackConfig(context.Context, uint64, map[string]interface{}, string) error
	ClearPackConfigCache(context.Context, []uint64, string) error
	ClearPackCategoryCache(context.Context) error
	ClearUserPackCache(context.Context, string) error
	ClearVendorPackCache(context.Context, string) error
	ClearPackCache(context.Context, uint64) error
	ClearPackShopCache(context.Context) error
	ActivatePacks(context.Context, []uint64, string) error
	DeactivatePacks(context.Context, []uint64, string) error
	DeletePackConfigs(context.Context, []uint64, string) error
	GeneratePackItemOdds(context.Context, []model.Item, int) (map[uint64]int, error)
}

type PackRepoImpl struct {
	db    *sqlx.DB
	cache *redis.Client
}

func NewPackRepo(db *sqlx.DB, cache *redis.Client) PackRepository {
	return &PackRepoImpl{db: db, cache: cache}
}

type PackError struct {
	message string
}

func (e *PackError) Error() string {
	return e.message
}

func (r *PackRepoImpl) CreatePackConfig(c context.Context, packConfig *model.PackConfig) (*model.PackConfig, error) {
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
		Insert(db.SCHEMA_PACK_CONFIGS).
		Columns(core.ModelColumns(packConfig)...).
		Values(core.StructValues(packConfig)...).
		Suffix("RETURNING \"id\"").
		ToSql()
	if err != nil {
		return nil, err
	}

	insertedId := uint64(0)
	err = tx.QueryRowContext(ctx, query, args...).Scan(&insertedId)
	if err != nil {
		return nil, err
	}

	// update vendor pack amount when pack is active
	if packConfig != nil && *packConfig.Active {
		if err = r.UpdateVendorPackAmount(ctx, *packConfig.VendorID, 1, tx); err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	packConfig, err = r.GetPackConfig(c, insertedId)
	if err != nil {
		return nil, err
	}
	return packConfig, nil
}

func (r *PackRepoImpl) AddPackItemConfigs(c context.Context, packItemConfigs []*model.PackItemConfig) error {
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
	query := psql.
		Insert(db.SCHEMA_PACK_ITEM_CONFIGS).
		Columns(core.ModelColumns(packItemConfigs[0])...)

	for _, config := range packItemConfigs {
		query = query.Values(core.StructValues(config)...)
	}
	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, sql, args...)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

// function that associates x amount of pack facts with a new owner and updates the current stock of packs
func (r *PackRepoImpl) BuyPacks(c context.Context, uid string, packConfig *model.PackConfig, amount float64, activeTokenRate *model.TokenCurrencyRate) (*model.PackBoughtResp, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
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

	// validate user has enough tokens to purchase amount of packs
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Select(fmt.Sprintf("t.balance >= %v*p.token_amount, p.token_amount, t.balance", amount)).
		From("financial.token_balance t").
		Join(fmt.Sprintf("main.pack_configs p on t.uid = '%v' and p.id = %v", uid, *packConfig.ID)).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	hasEnough := false
	tokenAmount := 0.0
	currBalance := 0.0
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&hasEnough, &tokenAmount, &currBalance); err != nil {
			return nil, err
		}
		if !hasEnough {
			err = &core.ErrorResp{Message: "User does not have sufficient token balance to buy this amount of packs"}
			return nil, err
		}
	}

	// get list of available packs
	psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err = psql.
		Select("id").
		From(db.SCHEMA_PACK_FACTS).
		Where(squirrel.And{squirrel.Eq{"pack_config_id": *packConfig.ID}, squirrel.Eq{"owner_id": nil}}).
		Limit(uint64(amount)).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err = tx.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	var packIds []uint64
	defer rows.Close()
	for rows.Next() {
		id := uint(0)
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		packIds = append(packIds, uint64(id))
	}

	// check if amount of packs is the same as amount wanting to purchase
	inStock := len(packIds)
	if inStock < int(amount) {
		if inStock == 0 {
			return nil, &core.ErrorResp{Message: "Sorry, there are currently no more of these packs available"}
		} else {
			return nil, &core.ErrorResp{Message: fmt.Sprintf("Sorry, there are only %v of these packs available in stock", inStock)}
		}
	}

	// associate packs with the user as the new owner
	now := time.Now().Format("2006-01-02 15:04:05")
	purchaseMap := map[string]interface{}{"owner_id": uid, "purchased_at": now}
	psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err = psql.
		Update(db.SCHEMA_PACK_FACTS).
		SetMap(purchaseMap).
		Where(squirrel.Eq{"id": packIds}).
		ToSql()
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		fmt.Println("Critical Error Buying Packs: ", err)
		return nil, &core.DBErrorResp{
			Message: "We are experiencing high demand for this pack and are unable to process your request at this time. Please try again later",
		}
	}

	// if err = r.AddPackOrder(ctx, now, uid, packConfig, packIds, tokenRateId, tx); err != nil {
	// 	return nil, err
	// }

	// adding the pack orderrs
	if packConfig.TokenAmount == nil {
		return nil, &core.ErrorResp{
			Message: "critical error: pack config does not have token amount",
		}
	}

	packOrders := make([]model.PackOrder, len(packIds))
	for i, id := range packIds {
		packId := id
		packOrder := model.PackOrder{
			PackId:      &packId,
			Uid:         &uid,
			OrderedAt:   &now,
			TokenAmount: &tokenAmount,
			TokenRateId: activeTokenRate.ID,
		}
		packOrders[i] = packOrder
	}
	packOrderQuery := psql.
		Insert(db.SCHEMA_PACK_ORDERS).
		Columns(core.ModelColumns(packOrders[0])...)

	for _, packOrder := range packOrders {
		packOrderQuery = packOrderQuery.Values(core.StructValues(packOrder)...)
	}
	queryStr, args, err := packOrderQuery.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(c, queryStr, args...)
	if err != nil {
		return nil, err
	}

	// update the users token balance
	newBalance := currBalance - (amount * tokenAmount)
	psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err = psql.
		Update(db.SCHEMA_TOKEN_BALANCE).
		Set("balance", newBalance).
		Where(squirrel.Eq{"uid": uid}).
		ToSql()
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &model.PackBoughtResp{PackIds: packIds, NewBalance: newBalance}, nil
}

func (r *PackRepoImpl) AddPackOrder(c context.Context, now string, uid string, packConfig *model.PackConfig, packIds []uint64, tokenRateId uint64, tx *sqlx.Tx) error {
	if packConfig.TokenAmount == nil {
		return &core.ErrorResp{
			Message: "critical error: pack config does not have token amount",
		}
	}
	packOrders := make([]model.PackOrder, len(packIds))
	tokenAmount := *packConfig.TokenAmount
	for i, id := range packIds {
		packId := id
		packOrder := model.PackOrder{
			PackId:      &packId,
			Uid:         &uid,
			OrderedAt:   &now,
			TokenAmount: &tokenAmount,
			TokenRateId: &tokenRateId,
		}
		packOrders[i] = packOrder
	}
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query := psql.
		Insert(db.SCHEMA_PACK_ORDERS).
		Columns(core.ModelColumns(packOrders[0])...)

	for _, packOrder := range packOrders {
		query = query.Values(core.StructValues(packOrder)...)
	}
	queryStr, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(c, queryStr, args...)
	if err != nil {
		return nil
	}

	return nil
}

func (r *PackRepoImpl) OpenPack(c context.Context, packId uint64, uid string) (*model.Pack, error) {
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
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Update(db.SCHEMA_PACK_FACTS).
		Set("opened_at", now).
		Where(squirrel.Eq{"id": packId}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := rows.RowsAffected()
	if rowsAffected > 0 {
		if err = tx.Commit(); err != nil {
			return nil, err
		}

		// remove cached user packs
		keys, err := r.cache.Keys(c, fmt.Sprintf("/user/packs/%v*", uid)).Result()
		if err != nil {
			return nil, err
		}
		for _, key := range keys {
			if err = r.cache.Del(c, key).Err(); err != nil {
				return nil, err
			}
		}

		pack, err := r.GetPack(c, packId)
		if err != nil {
			return nil, err
		}
		return pack, err
	}
	err = &core.ErrorResp{Message: "Pack fact record does not exist"}
	return nil, err
}

func (r *PackRepoImpl) GetPack(c context.Context, id uint64) (*model.Pack, error) {
	val, err := r.cache.Get(c, db.KEY_PACK+fmt.Sprintf("%v", id)).Result()
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
			Select(core.FlattenedPackFieldList...).
			From("main.pack_configs pc").
			Join("main.pack_facts p on pc.id = p.pack_config_id and p.active = true and p.opened_at is null").
			Join("main.pack_item_facts item_facts on p.id = item_facts.pack_id").
			Join("main.items i on item_facts.item_id = i.id").
			Where(squirrel.Eq{"p.id": id}).
			ToSql()
		if err != nil {
			return nil, err
		}
		fmt.Println(query)
		rows, err := tx.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}

		flattenedPacks := []model.PackFlatten{}
		defer rows.Close()
		for rows.Next() {
			flatPack := model.PackFlatten{}
			if err := rows.StructScan(&flatPack); err != nil {
				return nil, err
			}
			flattenedPacks = append(flattenedPacks, flatPack)
		}

		if len(flattenedPacks) <= 0 {
			return nil, &core.ErrorResp{Message: fmt.Sprintf("Pack fact with id %v does not exist or has already been opened", id)}
		} else {
			flattenedPack := flattenedPacks[0]
			items := []*model.Item{}
			pack := model.Pack{
				ID:           flattenedPack.PackFactId,
				PackConfigId: flattenedPack.PackConfigId,
				ImageUrl:     flattenedPack.ImageUrl,
				PackItemQty:  flattenedPack.PackItemQty,
				VendorId:     flattenedPack.VendorId,
				CreatedAt:    flattenedPack.CreatedAt,
				PurchasedAt:  flattenedPack.PurchasedAt,
				OpenedAt:     flattenedPack.OpenedAt,
				OwnerId:      flattenedPack.OwnerId,
				Active:       flattenedPack.Active,
				Description:  flattenedPack.Description,
				Title:        flattenedPack.Title,
				Items:        items,
			}
			for _, v := range flattenedPacks {
				item := model.Item{
					ID:              v.ItemId,
					VendorId:        v.ItemVendorId,
					ImageUrl:        v.ItemImageUrl,
					CreatedAt:       v.ItemCreatedAt,
					DeletedAt:       v.ItemDeletedAt,
					UpdatedAt:       v.ItemUpdatedAt,
					Description:     v.ItemDescription,
					Name:            v.ItemName,
					RarityId:        v.ItemRarityId,
					ContentMainUrl:  v.ItemContentMainUrl,
					ContentThumbUrl: v.ItemContentThumbUrl,
					ContentType:     v.ItemContentType,
					Active:          v.ItemActive,
					Notify:          v.ItemNotify,
				}
				pack.Items = append(pack.Items, &item)
			}

			packBytes, err := json.Marshal(pack)
			if err != nil {
				return nil, err
			}
			if err = r.cache.Set(c, db.KEY_PACK+fmt.Sprintf("%v", id), packBytes, time.Duration(3600)*time.Second).Err(); err != nil {
				return nil, err
			}
			return &pack, nil
		}
	} else {
		pack := model.Pack{}
		if err = json.Unmarshal([]byte(val), &pack); err != nil {
			return nil, err
		}
		return &pack, nil
	}
}

// func (r *PackRepoImpl) GetPackSortMappings(c context.Context) (map[string]string, error) {
// 	ctx, cancel := context.WithTimeout(c, 5*time.Second)
// 	defer cancel()

// 	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
// 	if err != nil {
// 		return nil, err
// 	}

// 	defer func() {
// 		if err = tx.Commit(); err != nil {
// 			fmt.Println(err)
// 		}
// 	}()

// 	category := "packs"
// 	return core.GetSortMappings(ctx, &category, tx, err)
// }

func (r *PackRepoImpl) GetPackConfig(c context.Context, packConfigId uint64) (*model.PackConfig, error) {
	val, err := r.cache.Get(c, db.KEY_PACK_CONFIG+fmt.Sprintf("%v", packConfigId)).Result()
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
			Select("*").
			From(db.SCHEMA_PACK_CONFIGS).
			Where(squirrel.Eq{"id": packConfigId}).
			ToSql()
		if err != nil {
			return nil, err
		}

		rows, err := tx.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}

		packConfig := model.PackConfig{}
		defer rows.Close()
		for rows.Next() {
			if err := rows.StructScan(&packConfig); err != nil {
				return &packConfig, err
			}
		}

		if packConfig.ID != nil {
			packConfigBytes, err := json.Marshal(packConfig)
			if err != nil {
				return nil, err
			}
			if err = r.cache.Set(c, db.KEY_PACK_CONFIG+fmt.Sprintf("%v", packConfigId), packConfigBytes, time.Duration(3600)*time.Second).Err(); err != nil {
				return nil, err
			}
		}
		return &packConfig, nil
	} else {
		packConfig := model.PackConfig{}
		if err = json.Unmarshal([]byte(val), &packConfig); err != nil {
			return nil, err
		}
		return &packConfig, nil
	}
}

func (r *PackRepoImpl) GetPackItems(c context.Context, packConfigId uint64) ([]*model.PackItemConfigExpanded, error) {
	val, err := r.cache.Get(c, db.KEY_PACK_ITEMS+fmt.Sprintf("%v", packConfigId)).Result()
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
			Select(
				"pic.qty as item_qty",
				"i.id",
				"i.created_at",
				"i.deleted_at",
				"i.updated_at",
				"i.description",
				"i.vendor_id",
				"i.name",
				"i.content_main_url",
				"i.content_thumb_url",
				"i.content_type",
				"i.value as item_value",
				"r.rarity",
			).
			From("main.pack_item_configs pic").
			Join("main.items i on i.id = pic.item_id").
			Join("main.rarity r on i.rarity_id = r.id").
			Where(squirrel.Eq{"pic.pack_config_id": packConfigId}).
			OrderBy("item_qty asc").
			ToSql()
		if err != nil {
			return nil, err
		}

		rows, err := tx.Queryx(query, args...)
		if err != nil {
			return nil, err
		}

		packItems := []*model.PackItemConfigExpanded{}
		defer rows.Close()
		for rows.Next() {
			packItem := model.PackItemConfigExpanded{}
			if err := rows.StructScan(&packItem); err != nil {
				return nil, err
			}
			packItems = append(packItems, &packItem)
		}

		packItemBytes, err := json.Marshal(packItems)
		if err != nil {
			return nil, err
		}
		if err = r.cache.Set(c, db.KEY_PACK_ITEMS+fmt.Sprintf("%v", packConfigId), packItemBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, err
		}
		return packItems, nil
	} else {
		packItems := []*model.PackItemConfigExpanded{}
		if err = json.Unmarshal([]byte(val), &packItems); err != nil {
			return nil, err
		}
		return packItems, nil
	}
}

func (r *PackRepoImpl) GetPackItemsPreview(c context.Context, packConfigId uint64) ([]*model.PackItemPreview, error) {
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

	rows, err := tx.QueryxContext(ctx, fmt.Sprintf(query.PackItemPreview, packConfigId))
	if err != nil {
		return nil, err
	}

	itemPreviews := []*model.PackItemPreview{}
	defer rows.Close()
	for rows.Next() {
		preview := model.PackItemPreview{}
		if err = rows.StructScan(&preview); err != nil {
			return nil, err
		}
		itemPreviews = append(itemPreviews, &preview)
	}

	return itemPreviews, nil
}

func (r *PackRepoImpl) AddPackCategories(c context.Context, categories []*model.PackCategory) error {
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
	query := psql.
		Insert(db.SCHEMA_PACK_CATEGORIES).
		Columns("pack_config_id", "category_id")

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
	return err
}

func (r *PackRepoImpl) GetUserPackAmount(c context.Context, uid string) (*uint64, error) {
	val, err := r.cache.Get(c, db.KEY_USER_PACK_AMOUNT+uid).Result()
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
			Select("count(*)").
			From("main.pack_facts").
			Where(squirrel.Eq{"owner_id": uid, "opened_at": nil}).
			ToSql()
		if err != nil {
			return nil, err
		}

		rows, err := tx.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		count := new(uint64)
		for rows.Next() {
			if err := rows.Scan(count); err != nil {
				return nil, err
			}
		}
		if err = r.cache.Set(c, db.KEY_USER_PACK_AMOUNT+uid, fmt.Sprintf("%v", count), time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, err
		}
		return count, nil
	} else {
		count := new(uint64)
		if err = json.Unmarshal([]byte(val), count); err != nil {
			return nil, err
		}
		return count, nil
	}
}

func (r *PackRepoImpl) GetPackItemConfigs(c context.Context, packConfigId uint64) ([]*model.PackItemConfig, error) {
	val, err := r.cache.Get(c, db.KEY_PACK_ITEM_CONFIGS+fmt.Sprintf("%v", packConfigId)).Result()
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
			Select("id", "pack_config_id", "item_id", "qty", "created_at", "removed_at").
			From(db.SCHEMA_PACK_ITEM_CONFIGS).
			Where(squirrel.Eq{"pack_config_id": packConfigId}).
			ToSql()
		if err != nil {
			return nil, err
		}

		rows, err := tx.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}

		packItemConfigs := []*model.PackItemConfig{}
		defer rows.Close()
		for rows.Next() {
			packItemConfig := model.PackItemConfig{}
			if err := rows.StructScan(&packItemConfig); err != nil {
				return nil, err
			}
			packItemConfigs = append(packItemConfigs, &packItemConfig)
		}

		packItemConfigBytes, err := json.Marshal(packItemConfigs)
		if err != nil {
			return nil, err
		}
		if err = r.cache.Set(c, db.KEY_PACK_ITEM_CONFIGS+fmt.Sprintf("%v", packConfigId), packItemConfigBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, err
		}
		return packItemConfigs, nil
	} else {
		packItemConfigs := []*model.PackItemConfig{}
		if err = json.Unmarshal([]byte(val), &packItemConfigs); err != nil {
			return nil, err
		}
		return packItemConfigs, nil
	}
}

func (r *PackRepoImpl) GetActivePackItems(c context.Context, itemIds []uint64) ([]string, []string, error) {
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

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Select("pc.title", "pc.vendor_id", "i.name").
		From("main.pack_configs pc").
		Join("main.pack_item_configs pic on pc.id = pic.pack_config_id and pc.active = true").
		Join("main.items i on i.id = pic.item_id").
		Where(squirrel.Eq{"i.id": itemIds}).
		ToSql()
	if err != nil {
		return nil, nil, err
	}

	rows, err := tx.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}

	itemNames := []string{}
	packTitles := []string{}

	defer rows.Close()
	for rows.Next() {
		packTitle := ""
		itemName := ""
		vendorId := ""
		if err = rows.Scan(&packTitle, &vendorId, &itemName); err != nil {
			return nil, nil, err
		}
		itemNames = append(itemNames, itemName)
		packTitles = append(packTitles, packTitle)
	}
	return itemNames, packTitles, nil
}

func (r *PackRepoImpl) GetPacksContainingItems(c context.Context, itemIds []uint64) ([]string, []string, error) {
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

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Select("pc.title", "pc.vendor_id", "i.name").
		From("main.pack_configs pc").
		Join("main.pack_item_configs pic on pc.id = pic.pack_config_id and pc.deleted_at is null").
		Join("main.items i on i.id = pic.item_id").
		Where(squirrel.Eq{"i.id": itemIds}).
		ToSql()
	if err != nil {
		return nil, nil, err
	}

	rows, err := tx.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}

	itemNames := []string{}
	packTitles := []string{}

	defer rows.Close()
	for rows.Next() {
		packTitle := ""
		itemName := ""
		vendorId := ""
		if err = rows.Scan(&packTitle, &vendorId, &itemName); err != nil {
			return nil, nil, err
		}
		itemNames = append(itemNames, itemName)
		packTitles = append(packTitles, packTitle)
	}
	return itemNames, packTitles, nil
}

func (r *PackRepoImpl) UploadPacks(c context.Context, packFacts []*model.PackFact, packConfigId uint64) ([]uint64, error) {
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
	query := psql.
		Insert(db.SCHEMA_PACK_FACTS).
		Columns(core.ModelColumns(packFacts[0])...).
		Suffix("RETURNING id")

	packIds := []uint64{}
	for _, packFact := range packFacts {
		query = query.Values(core.StructValues(packFact)...)
	}

	queryStr, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryxContext(ctx, queryStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		packIds = append(packIds, id)
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return packIds, nil
}

func (r *PackRepoImpl) UploadPackItems(c context.Context, packItemFacts []*model.PackItemFact) error {
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
	query := psql.
		Insert(db.SCHEMA_PACK_ITEM_FACTS).
		Columns(core.ModelColumns(packItemFacts[0])...)

	for _, packItemFact := range packItemFacts {
		query = query.Values(core.StructValues(packItemFact)...)
	}

	queryStr, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

func (r *PackRepoImpl) UpdatePackStock(c context.Context, packConfigId uint64, amount int, tx *sqlx.Tx) error {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Update(db.SCHEMA_PACK_CONFIGS).
		Set("current_stock", squirrel.Expr("current_stock + ?", amount)).
		Where(squirrel.Eq{"id": packConfigId}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(c, query, args...)
	if err != nil {
		return err
	}
	return nil
}

func (r *PackRepoImpl) UpdatePackQtySold(c context.Context, packConfigId uint64, amount int, tx *sqlx.Tx) error {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Update(db.SCHEMA_PACK_CONFIGS).
		Set("qty_sold", squirrel.Expr("qty_sold + ?", amount)).
		Where(squirrel.Eq{"id": packConfigId}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(c, query, args...)
	if err != nil {
		return err
	}
	return nil
}

func (r *PackRepoImpl) PatchPackConfig(c context.Context, packConfigId uint64, patchMap map[string]interface{}, vendorId string) error {
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
		Update(db.SCHEMA_PACK_CONFIGS).
		SetMap(patchMap).
		Where(squirrel.Eq{"id": packConfigId, "vendor_id": vendorId}).
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
		if err = tx.Commit(); err != nil {
			return err
		}
		return nil
	} else {
		err = &PackError{"Pack config either is not owned by the authorized user or could not be mapped using the given patch."}
		return err
	}
}

func (r *PackRepoImpl) PatchPackConfigs(c context.Context, packConfigIds []uint64, patchMap map[string]interface{}) error {
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
		Update(db.SCHEMA_PACK_CONFIGS).
		SetMap(patchMap).
		Where(squirrel.Eq{"id": packConfigIds}).
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

		return err
	} else {
		err = &PackError{
			fmt.Sprintf("An error occurred during the update. PackConfig either is not active or could not be mapped using the given patch. %v", err.Error()),
		}
		return err
	}
}

func (r *PackRepoImpl) DeleteInactivePacks(c context.Context, packConfigIds []uint64, tx *sqlx.Tx) error {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Delete(db.SCHEMA_PACK_FACTS).
		Where(squirrel.Eq{"pack_config_id": packConfigIds, "purchased_at": nil, "owner_id": nil}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(c, query, args...)
	return err
}

func (r *PackRepoImpl) ActivatePacks(c context.Context, packConfigIds []uint64, vendorId string) error {
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

	/* Activate packs only if they have not been deleted, previously inactive, and have not been scheduled yet */
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Update(db.SCHEMA_PACK_CONFIGS).
		SetMap(map[string]interface{}{
			"active": true,
		}).
		Where(squirrel.Eq{
			"id":         packConfigIds,
			"vendor_id":  vendorId,
			"active":     false,
			"deleted_at": nil,
		}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	// updating pack amount for all activated pack count
	if err = r.UpdateVendorPackAmount(ctx, vendorId, len(packConfigIds), tx); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *PackRepoImpl) DeactivatePacks(c context.Context, packConfigIds []uint64, vendorId string) error {
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

	/* Along with setting the active flag to false, we must also remove
	any previous release at date so it does not get picked up by pack release scheduling */
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Update(db.SCHEMA_PACK_CONFIGS).
		SetMap(map[string]interface{}{
			"active":     false,
			"release_at": nil,
		}).
		Where(squirrel.Eq{
			"id":        packConfigIds,
			"vendor_id": vendorId,
			"active":    true,
		}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	// updating pack amount for all deactivated pack count
	if err = r.UpdateVendorPackAmount(ctx, vendorId, -1*len(packConfigIds), tx); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *PackRepoImpl) DeletePackConfigs(c context.Context, packConfigIds []uint64, vendorId string) error {
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
		Update(db.SCHEMA_PACK_CONFIGS).
		SetMap(map[string]interface{}{"active": false, "deleted_at": time.Now().Format("2006-01-02 15:04:05")}).
		Where(squirrel.Eq{"id": packConfigIds, "vendor_id": vendorId}).
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
		// update vendor pack amount
		if err = r.UpdateVendorPackAmount(ctx, vendorId, len(packConfigIds)*-1, tx); err != nil {
			return err
		}

		// remove all packs in market that are now inactive
		if err = r.DeleteInactivePacks(ctx, packConfigIds, tx); err != nil {
			return err
		}

		if err = tx.Commit(); err != nil {
			return err
		}
		return nil
	} else {
		err = &ItemError{"An error occurred deleting pack config. Pack config either does not exist or cannot be deleted"}
		return err
	}
}

func (r *PackRepoImpl) UpdateVendorPackAmount(c context.Context, vendorId string, packAmount int, tx *sqlx.Tx) error {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Update(db.SCHEMA_VENDORS).
		Set("pack_amount", squirrel.Expr("GREATEST(0, pack_amount + ?)", packAmount)).
		Where(squirrel.Eq{"uid": vendorId}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(c, query, args...)
	if err != nil {
		return err
	}
	return nil
}

func (r *PackRepoImpl) ClearPackConfigCache(c context.Context, packConfigIds []uint64, vendorId string) error {
	allKeys := []string{}

	// packs in shop
	keys, err := r.cache.Keys(c, "/vendors/packs*").Result()
	if err != nil {
		return err
	}
	allKeys = append(allKeys, keys...)

	// vendor profile pack entries
	keys, err = r.cache.Keys(c, "/vendor/packs/profile/"+vendorId+"*").Result()
	if err != nil {
		return err
	}
	allKeys = append(allKeys, keys...)

	// vendor pack list
	keys, err = r.cache.Keys(c, "/vendor/packs*").Result()
	if err != nil {
		return err
	}
	allKeys = append(allKeys, keys...)

	// pack configs
	for _, id := range packConfigIds {
		keys, err = r.cache.Keys(c, fmt.Sprintf("%v%v", db.KEY_PACK_CONFIG, id)).Result()
		if err != nil {
			return err
		}
	}
	allKeys = append(allKeys, keys...)

	// invalidate all cache entries
	return r.cache.Del(c, allKeys...).Err()
}

func (r *PackRepoImpl) ClearPackCategoryCache(c context.Context) error {
	if err := r.cache.Del(c, db.KEY_CATEGORY+"pack_categories").Err(); err != nil {
		return err
	}
	return nil
}

func (r *PackRepoImpl) ClearPackShopCache(c context.Context) error {
	keysToDelete := []string{
		"/vendors/packs*",
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

func (r *PackRepoImpl) ClearUserPackCache(c context.Context, uid string) error {
	keys, err := r.cache.Keys(c, fmt.Sprintf("/user/packs/%v*", uid)).Result()
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

func (r *PackRepoImpl) ClearVendorPackCache(c context.Context, vendorId string) error {
	keysToDelete := []string{
		fmt.Sprintf("/vendor/packs/%v*", vendorId),
		fmt.Sprintf("/vendor/packs/profile/%v*", vendorId),
		"/vendors/packs*",
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

func (r *PackRepoImpl) GeneratePackItemOdds(c context.Context, items []model.Item, totalItems int) (map[uint64]int, error) {
	return core.GenerateOdds(items, totalItems)
}

func (r *PackRepoImpl) ClearPackCache(c context.Context, packId uint64) error {
	return r.cache.Del(c, db.KEY_PACK+fmt.Sprintf("%v", packId)).Err()
}
