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

	"github.com/redis/go-redis/v9"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type VendorRepository interface {
	GetVendorsPage(context.Context, uint64, string, string, string, string) ([]*model.PageVendor, *uint64, error)
	GetVendorShopPackPage(context.Context, uint64, string, string, string, string) ([]*model.PageVendorPack, *uint64, error)
	GetVendorProfilePackPage(context.Context, string, uint64, string, string, string) ([]*model.PageVendorPack, *uint64, error)
	GetVendorPackListPage(context.Context, string, uint64, string, string, string, string) ([]*model.PageVendorPack, *uint64, error)
	GetVendorItemListPage(context.Context, string, uint64, string, string, string, string) ([]*model.PageVendorItem, *uint64, error)
	GetVendorSortList(context.Context, string) ([]string, error)
	GetVendorSortMappings(context.Context) (map[string]string, error)
	GetVendorShopPackSortMappings(context.Context) (map[string]string, error)
	GetVendorPackSortMappings(context.Context) (map[string]string, error)
	GetVendorItemSortMappings(context.Context) (map[string]string, error)
	GetVendorAmount(context.Context) (*uint64, error)
	GetActiveVendor(context.Context, string) (*model.Vendor, error)
	GetVendor(context.Context, string) (*model.Vendor, error)
	GetVendorPackCategories(context.Context) ([]string, error)
	GetVendorItemCategories(context.Context) ([]string, error)
	GetVendorCategories(context.Context, string) ([]*model.VendorCategoryExpanded, error)
	AddVendorCategories(context.Context, []*model.VendorCategory) error
	RemoveVendorCategories(context.Context, string, *sqlx.Tx) error
	ApproveVendor(context.Context, string) (*model.Vendor, error)
	RemoveVendor(context.Context, string) error
	PatchVendor(context.Context, string, map[string]interface{}) error
	ClearVendorCache(context.Context, string) error
}

type VendorRepoImpl struct {
	db    *sqlx.DB
	cache *redis.Client
}

func NewVendorRepo(db *sqlx.DB, cache *redis.Client) VendorRepository {
	return &VendorRepoImpl{db: db, cache: cache}
}

func (r *VendorRepoImpl) GetVendorsPage(
	c context.Context, pageNumber uint64, sortBy string, filterOn string, searchStr string, urlPath string) ([]*model.PageVendor, *uint64, error) {

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

		pageSizeStr := os.Getenv("VENDOR_PAGE_SIZE")
		pageSize, err := strconv.ParseUint(pageSizeStr, 10, 64)
		if err != nil {
			return nil, nil, err
		}
		var rows *sqlx.Rows
		if sortBy != "" {
			if filterOn != "" {
				rows, err = tx.QueryxContext(ctx, fmt.Sprintf(
					query.VendorPageSortByFilterOn,
					filterOn,
					searchStr,
					searchStr,
					sortBy,
					pageSize,
					(pageSize)*(pageNumber),
				))
			} else {
				rows, err = tx.QueryxContext(ctx, fmt.Sprintf(
					query.VendorPageSortBy,
					searchStr,
					searchStr,
					sortBy,
					pageSize,
					(pageSize)*(pageNumber),
				))
			}
		} else if filterOn != "" {
			rows, err = tx.QueryxContext(ctx, fmt.Sprintf(
				query.VendorPageFilterOn,
				filterOn,
				searchStr,
				searchStr,
				pageSize,
				(pageSize)*(pageNumber),
			))
		} else {
			rows, err = tx.QueryxContext(ctx, fmt.Sprintf(
				query.VendorPageDefault,
				searchStr,
				searchStr,
				pageSize,
				(pageSize)*(pageNumber),
			))
		}
		if err != nil {
			return nil, nil, err
		}

		vendors := []*model.PageVendor{}
		defer rows.Close()
		for rows.Next() {
			vendor := model.PageVendor{}
			if err := rows.StructScan(&vendor); err != nil {
				return nil, nil, err
			}
			if vendor.RawCategories != nil {
				categories := strings.Split(*vendor.RawCategories, ",")
				vendor.Categories = categories
			}
			vendors = append(vendors, &vendor)
		}
		vendorAmount := new(uint64)
		if len(vendors) <= 0 {
			*vendorAmount = 0
		} else {
			vendorAmount = vendors[len(vendors)-1].VendorAmount
		}

		vendorBytes, err := json.Marshal(vendors)
		if err != nil {
			return nil, nil, err
		}
		err = r.cache.Set(c, urlPath, vendorBytes, time.Duration(3600)*time.Second).Err()
		if err != nil {
			return nil, nil, err
		}
		return vendors, vendorAmount, nil
	} else {
		vendors := []*model.PageVendor{}
		if err = json.Unmarshal([]byte(val), &vendors); err != nil {
			return nil, nil, err
		}
		vendorAmount := new(uint64)
		if len(vendors) <= 0 {
			*vendorAmount = 0
		} else {
			vendorAmount = vendors[len(vendors)-1].VendorAmount
		}
		return vendors, vendorAmount, nil
	}
}

func (r *VendorRepoImpl) GetVendorShopPackPage(
	c context.Context, pageNumber uint64, sortBy string, filterOn string, searchStr string, urlPath string) ([]*model.PageVendorPack, *uint64, error) {

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

		pageSizeStr := os.Getenv("VENDOR_PACK_SHOP_PAGE_SIZE")
		pageSize, err := strconv.ParseUint(pageSizeStr, 10, 64)
		if err != nil {
			return nil, nil, err
		}

		var rows *sqlx.Rows
		if sortBy != "" {
			if filterOn != "" {
				rows, err = r.db.QueryxContext(ctx, fmt.Sprintf(
					query.ShopPackPageSortByFilterOn,
					filterOn,
					searchStr,
					searchStr,
					searchStr,
					sortBy,
					pageSize,
					(pageSize)*(pageNumber)),
				)
			} else {
				rows, err = r.db.QueryxContext(ctx, fmt.Sprintf(
					query.ShopPackPageSortBy,
					searchStr,
					searchStr,
					searchStr,
					sortBy,
					pageSize,
					(pageSize)*(pageNumber)),
				)
			}
		} else if filterOn != "" {
			rows, err = r.db.QueryxContext(ctx, fmt.Sprintf(
				query.ShopPackPageFilterOn,
				filterOn,
				searchStr,
				searchStr,
				searchStr,
				pageSize,
				(pageSize)*(pageNumber)),
			)
		} else {
			rows, err = r.db.QueryxContext(ctx, fmt.Sprintf(
				query.ShopPackPageDefault,
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

		packs := []*model.PageVendorPack{}
		defer rows.Close()
		for rows.Next() {
			pack := model.PageVendorPack{}
			if err := rows.StructScan(&pack); err != nil {
				return nil, nil, err
			}
			if pack.RawCategories != nil {
				categories := strings.Split(*pack.RawCategories, ",")
				pack.Categories = categories
			}
			packs = append(packs, &pack)
		}
		packAmount := new(uint64)
		if len(packs) <= 0 {
			*packAmount = 0
		} else {
			packAmount = packs[len(packs)-1].TotalPacksAmount
		}

		packBytes, err := json.Marshal(packs)
		if err != nil {
			return nil, nil, err
		}
		if err = r.cache.Set(c, urlPath, packBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, nil, err
		}
		return packs, packAmount, nil
	} else {
		packs := []*model.PageVendorPack{}
		if err = json.Unmarshal([]byte(val), &packs); err != nil {
			return nil, nil, err
		}
		packAmount := new(uint64)
		if len(packs) <= 0 {
			*packAmount = 0
		} else {
			packAmount = packs[len(packs)-1].TotalPacksAmount
		}
		return packs, packAmount, nil
	}
}

func (r *VendorRepoImpl) GetVendorProfilePackPage(
	c context.Context, vendorId string, pageNumber uint64, sortBy string, filterOn string, urlPath string) ([]*model.PageVendorPack, *uint64, error) {

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

		pageSizeStr := os.Getenv("VENDOR_PROFILE_PACK_PAGE_SIZE")
		pageSize, err := strconv.ParseUint(pageSizeStr, 10, 64)
		if err != nil {
			return nil, nil, err
		}

		var rows *sqlx.Rows
		if sortBy != "" {
			if filterOn != "" {
				rows, err = r.db.QueryxContext(ctx, fmt.Sprintf(query.VendorProfilePackPageSortByFilterOn, vendorId, filterOn, sortBy, pageSize, (pageSize)*(pageNumber)))
			} else {
				rows, err = r.db.QueryxContext(ctx, fmt.Sprintf(query.VendorProfilePackPageSortBy, vendorId, sortBy, pageSize, (pageSize)*(pageNumber)))
			}
		} else if filterOn != "" {
			rows, err = r.db.QueryxContext(ctx, fmt.Sprintf(query.VendorProfilePackPageFilterOn, vendorId, filterOn, pageSize, (pageSize)*(pageNumber)))
		} else {
			rows, err = r.db.QueryxContext(ctx, fmt.Sprintf(query.VendorProfilePackPageDefault, vendorId, pageSize, (pageSize)*(pageNumber)))
		}
		if err != nil {
			return nil, nil, err
		}

		packs := []*model.PageVendorPack{}
		defer rows.Close()
		for rows.Next() {
			pack := model.PageVendorPack{}
			if err := rows.StructScan(&pack); err != nil {
				return nil, nil, err
			}
			if pack.RawCategories != nil {
				categories := strings.Split(*pack.RawCategories, ",")
				pack.Categories = categories
			}
			packs = append(packs, &pack)
		}
		packAmount := new(uint64)
		if len(packs) <= 0 {
			*packAmount = 0
		} else {
			packAmount = packs[len(packs)-1].TotalPacksAmount
		}

		packBytes, err := json.Marshal(packs)
		if err != nil {
			return nil, nil, err
		}
		if err = r.cache.Set(c, urlPath, packBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, nil, err
		}
		return packs, packAmount, nil
	} else {
		packs := []*model.PageVendorPack{}
		if err = json.Unmarshal([]byte(val), &packs); err != nil {
			return nil, nil, err
		}
		packAmount := new(uint64)
		if len(packs) <= 0 {
			*packAmount = 0
		} else {
			packAmount = packs[len(packs)-1].TotalPacksAmount
		}
		return packs, packAmount, nil
	}
}

func (r *VendorRepoImpl) GetVendorPackListPage(
	c context.Context, vendorId string, pageNumber uint64, sortBy string, filterOn string, searchStr string, urlPath string) ([]*model.PageVendorPack, *uint64, error) {

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

		pageSizeStr := os.Getenv("VENDOR_PACKS_LIST_PAGE_SIZE")
		pageSize, err := strconv.ParseUint(pageSizeStr, 10, 64)
		if err != nil {
			return nil, nil, err
		}

		var rows *sqlx.Rows
		if sortBy != "" {
			if filterOn != "" {
				fmt.Printf(query.VendorPackListPageSortByFilterOn, vendorId, filterOn, sortBy, pageSize, (pageSize)*(pageNumber))
				rows, err = r.db.QueryxContext(ctx, fmt.Sprintf(
					query.VendorPackListPageSortByFilterOn,
					vendorId,
					filterOn,
					searchStr,
					searchStr,
					sortBy,
					pageSize,
					(pageSize)*(pageNumber)),
				)
			} else {
				rows, err = r.db.QueryxContext(ctx, fmt.Sprintf(
					query.VendorPackListPageSortBy,
					vendorId,
					searchStr,
					searchStr,
					sortBy,
					pageSize,
					(pageSize)*(pageNumber)),
				)
			}
		} else if filterOn != "" {
			rows, err = r.db.QueryxContext(ctx, fmt.Sprintf(
				query.VendorPackListPageFilterOn,
				vendorId,
				filterOn,
				searchStr,
				searchStr,
				pageSize,
				(pageSize)*(pageNumber)),
			)
		} else {
			rows, err = r.db.QueryxContext(ctx, fmt.Sprintf(
				query.VendorPackListPageDefault,
				vendorId,
				searchStr,
				searchStr,
				pageSize,
				(pageSize)*(pageNumber)),
			)
		}
		if err != nil {
			return nil, nil, err
		}

		packs := []*model.PageVendorPack{}
		defer rows.Close()
		for rows.Next() {
			pack := model.PageVendorPack{}
			if err := rows.StructScan(&pack); err != nil {
				return nil, nil, err
			}
			if pack.RawCategories != nil {
				categories := strings.Split(*pack.RawCategories, ",")
				pack.Categories = categories
			}
			packs = append(packs, &pack)
		}
		packAmount := new(uint64)
		if len(packs) <= 0 {
			*packAmount = 0
		} else {
			packAmount = packs[len(packs)-1].TotalPacksAmount
		}

		packBytes, err := json.Marshal(packs)
		if err != nil {
			return nil, nil, err
		}
		if err = r.cache.Set(c, urlPath, packBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, nil, err
		}
		return packs, packAmount, nil
	} else {
		packs := []*model.PageVendorPack{}
		if err = json.Unmarshal([]byte(val), &packs); err != nil {
			return nil, nil, err
		}
		packAmount := new(uint64)
		if len(packs) <= 0 {
			*packAmount = 0
		} else {
			packAmount = packs[len(packs)-1].TotalPacksAmount
		}
		return packs, packAmount, nil
	}
}

func (r *VendorRepoImpl) GetVendorItemListPage(
	c context.Context, vendorId string, pageNumber uint64, sortBy string, filterOn string, searchStr string, urlPath string) ([]*model.PageVendorItem, *uint64, error) {

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

		pageSizeStr := os.Getenv("VENDOR_ITEM_LIST_PAGE_SIZE")
		pageSize, err := strconv.ParseUint(pageSizeStr, 10, 64)
		if err != nil {
			return nil, nil, err
		}

		var rows *sqlx.Rows
		fmt.Println("sort by: ", sortBy)
		fmt.Println("filter on: ", filterOn)
		if sortBy != "" {
			if filterOn != "" {
				rows, err = r.db.QueryxContext(ctx, fmt.Sprintf(
					query.VendorItemListPageSortByFilterOn,
					vendorId,
					filterOn,
					searchStr,
					searchStr,
					sortBy,
					pageSize,
					(pageSize)*(pageNumber)),
				)
			} else {
				rows, err = r.db.QueryxContext(ctx, fmt.Sprintf(
					query.VendorItemListPageSortBy,
					vendorId,
					searchStr,
					searchStr,
					sortBy,
					pageSize,
					(pageSize)*(pageNumber)),
				)
			}
		} else if filterOn != "" {
			rows, err = r.db.QueryxContext(ctx, fmt.Sprintf(
				query.VendorItemListPageFilterOn,
				vendorId,
				filterOn,
				searchStr,
				searchStr,
				pageSize,
				(pageSize)*(pageNumber)),
			)
		} else {
			rows, err = r.db.QueryxContext(ctx, fmt.Sprintf(
				query.VendorItemListPageDefault,
				vendorId,
				searchStr,
				searchStr,
				pageSize,
				(pageSize)*(pageNumber)),
			)
		}
		if err != nil {
			return nil, nil, err
		}

		items := []*model.PageVendorItem{}
		defer rows.Close()
		for rows.Next() {
			item := model.PageVendorItem{}
			if err := rows.StructScan(&item); err != nil {
				return nil, nil, err
			}
			if item.RawCategories != nil {
				categories := strings.Split(*item.RawCategories, ",")
				item.Categories = categories
			}
			items = append(items, &item)
		}
		itemAmount := new(uint64)
		if len(items) <= 0 {
			*itemAmount = 0
		} else {
			itemAmount = items[len(items)-1].TotalItemsAmount
		}

		itemBytes, err := json.Marshal(items)
		if err != nil {
			return nil, nil, err
		}
		if err = r.cache.Set(c, urlPath, itemBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, nil, err
		}
		return items, itemAmount, nil
	} else {
		items := []*model.PageVendorItem{}
		if err = json.Unmarshal([]byte(val), &items); err != nil {
			return nil, nil, err
		}
		itemAmount := new(uint64)
		if len(items) <= 0 {
			*itemAmount = 0
		} else {
			itemAmount = items[len(items)-1].TotalItemsAmount
		}
		return items, itemAmount, nil
	}
}

func (r *VendorRepoImpl) GetVendorSortList(c context.Context, urlPath string) ([]string, error) {
	val, err := r.cache.Get(c, urlPath).Result()
	if err != nil {
		list, err := core.GetSortList("vendors")
		if err != nil {
			return nil, err
		}
		listBytes, err := json.Marshal(list)
		if err != nil {
			return nil, err
		}
		if err = r.cache.Set(c, urlPath, listBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, err
		}
		return list, err
	} else {
		list := []string{}
		if err = json.Unmarshal([]byte(val), &list); err != nil {
			return nil, err
		}
		return list, nil
	}
}

func (r *VendorRepoImpl) GetVendorSortMappings(c context.Context) (map[string]string, error) {
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
	category := "vendors"
	return core.GetSortMappings(ctx, &category, tx, err)
}

func (r *VendorRepoImpl) GetVendorShopPackSortMappings(c context.Context) (map[string]string, error) {
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
	category := "packs"
	return core.GetSortMappings(ctx, &category, tx, err)
}

func (r *VendorRepoImpl) GetVendorPackSortMappings(c context.Context) (map[string]string, error) {
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
	category := "vendor packs"
	return core.GetSortMappings(ctx, &category, tx, err)
}

func (r *VendorRepoImpl) GetVendorItemSortMappings(c context.Context) (map[string]string, error) {
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
	category := "items"
	return core.GetSortMappings(ctx, &category, tx, err)
}

func (r *VendorRepoImpl) GetVendorAmount(c context.Context) (*uint64, error) {
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
		From(db.SCHEMA_VENDORS).
		Where(squirrel.Eq{"active": true}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var count uint64
		if err := rows.Scan(&count); err != nil {
			return nil, err
		}
		return &count, nil
	}
	err = &core.ErrorResp{Message: "Unable to get count of vendors"}
	return nil, err
}

func (r *VendorRepoImpl) GetActiveVendor(c context.Context, vendorId string) (*model.Vendor, error) {
	val, err := r.cache.Get(c, db.KEY_VENDOR+vendorId).Result()
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
			Select("v.*").
			From("main.vendors v").
			Where(squirrel.Eq{"v.active": true, "v.uid": vendorId}).
			ToSql()
		if err != nil {
			return nil, err
		}

		rows, err := tx.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}

		vendor := model.Vendor{}
		defer rows.Close()
		for rows.Next() {
			if err := rows.StructScan(&vendor); err != nil {
				return nil, err
			}
		}

		vendorBytes, err := json.Marshal(vendor)
		if err != nil {
			return nil, err
		}
		if err = r.cache.Set(c, db.KEY_VENDOR+vendorId, vendorBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, err
		}

		return &vendor, nil
	} else {
		vendor := model.Vendor{}
		if err = json.Unmarshal([]byte(val), &vendor); err != nil {
			return nil, err
		}
		return &vendor, nil
	}
}

func (r *VendorRepoImpl) GetVendor(c context.Context, vendorId string) (*model.Vendor, error) {
	val, err := r.cache.Get(c, db.KEY_VENDOR+vendorId).Result()
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
			Select("v.*").
			From("main.vendors v").
			Where(squirrel.Eq{"v.uid": vendorId}).
			ToSql()
		if err != nil {
			return nil, err
		}

		rows, err := tx.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}

		vendor := model.Vendor{}
		defer rows.Close()
		for rows.Next() {
			if err := rows.StructScan(&vendor); err != nil {
				return nil, err
			}
		}

		vendorBytes, err := json.Marshal(vendor)
		if err != nil {
			return nil, err
		}
		if err = r.cache.Set(c, db.KEY_VENDOR+vendorId, vendorBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, err
		}

		return &vendor, nil
	} else {
		vendor := model.Vendor{}
		if err = json.Unmarshal([]byte(val), &vendor); err != nil {
			return nil, err
		}
		return &vendor, nil
	}
}

func (r *VendorRepoImpl) GetVendorCategories(c context.Context, vendorId string) ([]*model.VendorCategoryExpanded, error) {
	val, err := r.cache.Get(c, db.KEY_VENDOR+vendorId).Result()
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
			Select("vc.vendor_id", "vc.category_id", "c.category").
			From("main.categories c").
			Join(fmt.Sprintf("main.vendor_categories vc on c.id = vc.category_id and vc.vendor_id = '%v'", vendorId)).
			ToSql()
		if err != nil {
			return nil, err
		}

		fmt.Println("QUERY: ", query)
		rows, err := tx.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}

		categories := []*model.VendorCategoryExpanded{}
		defer rows.Close()
		for rows.Next() {
			category := model.VendorCategoryExpanded{}
			if err = rows.StructScan(&category); err != nil {
				return nil, err
			}
			categories = append(categories, &category)
		}

		if len(categories) <= 0 {
			return categories, nil
		}

		categoryBytes, err := json.Marshal(categories)
		if err != nil {
			return nil, err
		}
		if err = r.cache.Set(c, db.KEY_VENDOR_CATEGORIES+vendorId, categoryBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, err
		}
		return categories, nil
	} else {
		categories := []*model.VendorCategoryExpanded{}
		if err = json.Unmarshal([]byte(val), &categories); err != nil {
			return nil, err
		}
		return categories, nil
	}
}

func (r *VendorRepoImpl) AddVendorCategories(c context.Context, categories []*model.VendorCategory) error {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}

	defer func() {
		if err = tx.Commit(); err != nil {
			fmt.Println(err)
		}
	}()

	// removing any existing vendor categories within transaction
	vendorId := categories[0].VendorId
	if err = r.RemoveVendorCategories(ctx, *vendorId, tx); err != nil {
		return err
	}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query := psql.
		Insert(db.SCHEMA_VENDOR_CATEGORIES).
		Columns("vendor_id", "category_id")

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

func (r *VendorRepoImpl) RemoveVendorCategories(c context.Context, vendorId string, tx *sqlx.Tx) error {
	// remove vendor categories cache
	if err := r.ClearVendorCache(c, vendorId); err != nil {
		return err
	}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Delete(db.SCHEMA_VENDOR_CATEGORIES).
		Where(squirrel.Eq{"vendor_id": vendorId}).
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

func (r *VendorRepoImpl) GetVendorPackCategories(c context.Context) ([]string, error) {
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

func (r *VendorRepoImpl) GetVendorItemCategories(c context.Context) ([]string, error) {
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

	return core.GetCategories(ctx, "item_categories", tx, err)
}

func (r *VendorRepoImpl) ApproveVendor(c context.Context, uid string) (*model.Vendor, error) {
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
		Update(db.SCHEMA_USERS).
		Set("is_vendor", true).
		Where(squirrel.Eq{"uid": uid}).
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
		vendor, err := r.GetActiveVendor(ctx, uid)
		if err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return vendor, nil
	}
	return nil, &core.ErrorResp{Message: "could not approve vendor. User either does not exist or an unexpected error occurred when updating"}
}

func (r *VendorRepoImpl) RemoveVendor(c context.Context, uid string) error {
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
		Set("is_vendor", false).
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
		if err := tx.Commit(); err != nil {
			return err
		}
		return nil
	}
	return &core.ErrorResp{Message: "could not remove vendor. User either does not exist or an unexpected error occurred when updating"}
}

func (r *VendorRepoImpl) PatchVendor(c context.Context, vendorId string, dbPatchMap map[string]interface{}) error {
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
		Update(db.SCHEMA_VENDORS).
		SetMap(dbPatchMap).
		Where(squirrel.Eq{"uid": vendorId}).
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

	err = &UserError{message: fmt.Sprintf("Vendor with vendorId: %v does not exist", vendorId)}
	return err
}

func (r *VendorRepoImpl) UpdateVendorPackAmount(c context.Context, packAmount int64, vendorId string) error {
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
		Update(db.SCHEMA_VENDORS).
		Set("pack_amount", squirrel.Expr("pack_amount + ?", packAmount)).
		Where(squirrel.Eq{"uid": vendorId}).
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
	return &UserError{message: fmt.Sprintf("An error occurred updating the pack amount for vendor ID: %v", vendorId)}
}

func (r *VendorRepoImpl) ClearVendorCache(c context.Context, vendorId string) error {
	keysToDelete := []string{
		db.KEY_VENDOR + vendorId,
		"vendor_categories",
		db.KEY_VENDOR_CATEGORIES + vendorId,
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
