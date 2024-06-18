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

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type AnalyticsRepository interface {
	TotalVendorPacksSold(context.Context, string) (*uint64, error)
	TotalRevenueGenerated(context.Context, string) (*float64, error)
	FavoriteAmount(context.Context, string) (*uint64, error)
	AvgPackQtyPurchased(context.Context, string) (*float64, error)
	MinPackQtyPurchased(context.Context, string) (*uint64, error)
	MaxPackQtyPurchased(context.Context, string) (*uint64, error)
	TopCustomers(context.Context, string) ([]*model.CustomerAnalytic, error)
	PackSales(context.Context, string, string, int64, int64, string) ([]*model.PackSalesAnalytic, error)
	PackQtySold(context.Context, string) ([]*model.PackQtySoldAnalytic, error)
}

type AnalyticsRepoImpl struct {
	db    *sqlx.DB
	cache *redis.Client
}

func NewAnalyticsRepo(db *sqlx.DB, cache *redis.Client) AnalyticsRepository {
	return &AnalyticsRepoImpl{db: db, cache: cache}
}

func (r *AnalyticsRepoImpl) TotalVendorPacksSold(c context.Context, vendorId string) (*uint64, error) {
	val, err := r.cache.Get(c, db.KEY_TOTAL_PACKS_SOLD+vendorId).Result()
	if err != nil {
		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()

		tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return nil, err
		}

		defer func() {
			tx.Commit()
		}()

		rows, err := tx.QueryxContext(
			ctx,
			fmt.Sprintf(query.TotalVendorPacksSold, vendorId),
		)

		totalAmount := new(uint64)
		defer rows.Close()
		for rows.Next() {
			if err = rows.Scan(totalAmount); err != nil {
				return nil, err
			}
		}
		if totalAmount == nil {
			return nil, &core.ErrorResp{Message: "Error, total amount is null"}
		}

		totalAmountBytes, err := json.Marshal(totalAmount)
		if err != nil {
			return nil, err
		}
		if err = r.cache.Set(c, db.KEY_TOTAL_PACKS_SOLD+vendorId, totalAmountBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, err
		}
		return totalAmount, nil
	} else {
		totalAmount := new(uint64)
		if err = json.Unmarshal([]byte(val), totalAmount); err != nil {
			return nil, err
		}
		return totalAmount, nil
	}
}

func (r *AnalyticsRepoImpl) TotalRevenueGenerated(c context.Context, vendorId string) (*float64, error) {
	val, err := r.cache.Get(c, db.KEY_TOTAL_REVENUE+vendorId).Result()
	if err != nil {
		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()

		tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return nil, err
		}

		defer func() {
			tx.Commit()
		}()

		rows, err := tx.QueryxContext(
			ctx,
			fmt.Sprintf(query.TotalRevenueGenerated, vendorId),
		)

		totalRevenue := new(float64)
		defer rows.Close()
		for rows.Next() {
			if err = rows.Scan(totalRevenue); err != nil {
				return nil, err
			}
		}
		if totalRevenue == nil {
			return nil, &core.ErrorResp{Message: "Error, total revenue is null"}
		}

		totalRevenueBytes, err := json.Marshal(totalRevenue)
		if err != nil {
			return nil, err
		}
		if err = r.cache.Set(c, db.KEY_TOTAL_REVENUE+vendorId, totalRevenueBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, err
		}
		return totalRevenue, nil
	} else {
		totalRevenue := new(float64)
		if err = json.Unmarshal([]byte(val), totalRevenue); err != nil {
			return nil, err
		}
		return totalRevenue, nil
	}
}

func (r *AnalyticsRepoImpl) FavoriteAmount(c context.Context, vendorId string) (*uint64, error) {
	val, err := r.cache.Get(c, db.KEY_FAVORITE_AMOUNT+vendorId).Result()
	if err != nil {
		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()

		tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return nil, err
		}

		defer func() {
			tx.Commit()
		}()

		rows, err := tx.QueryxContext(
			ctx,
			fmt.Sprintf(query.FavoritesAmount, vendorId),
		)

		favoriteAmount := new(uint64)
		defer rows.Close()
		for rows.Next() {
			if err = rows.Scan(favoriteAmount); err != nil {
				return nil, err
			}
		}
		if favoriteAmount == nil {
			return nil, &core.ErrorResp{Message: "Error, no such vendor exists"}
		}

		favoriteAmountBytes, err := json.Marshal(favoriteAmount)
		if err != nil {
			return nil, err
		}
		if err = r.cache.Set(c, db.KEY_FAVORITE_AMOUNT+vendorId, favoriteAmountBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, err
		}
		return favoriteAmount, nil
	} else {
		favoriteAmount := new(uint64)
		if err = json.Unmarshal([]byte(val), favoriteAmount); err != nil {
			return nil, err
		}
		return favoriteAmount, nil
	}
}

func (r *AnalyticsRepoImpl) AvgPackQtyPurchased(c context.Context, vendorId string) (*float64, error) {
	val, err := r.cache.Get(c, db.KEY_AVG_PURCHASED_QTY+vendorId).Result()
	if err != nil {
		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()

		tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return nil, err
		}

		defer func() {
			tx.Commit()
		}()

		rows, err := tx.QueryxContext(
			ctx,
			fmt.Sprintf(query.AveragePackQtyPurchased, vendorId),
		)

		avgPurchased := new(float64)
		defer rows.Close()
		for rows.Next() {
			if err = rows.Scan(avgPurchased); err != nil {
				return nil, err
			}
		}
		if avgPurchased == nil {
			return nil, &core.ErrorResp{Message: "Error, avg purchased qty is null"}
		}

		avgPurchasedBytes, err := json.Marshal(avgPurchased)
		if err != nil {
			return nil, err
		}
		if err = r.cache.Set(c, db.KEY_AVG_PURCHASED_QTY+vendorId, avgPurchasedBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, err
		}
		return avgPurchased, nil
	} else {
		avgPurchased := new(float64)
		if err = json.Unmarshal([]byte(val), avgPurchased); err != nil {
			return nil, err
		}
		return avgPurchased, nil
	}
}

func (r *AnalyticsRepoImpl) MinPackQtyPurchased(c context.Context, vendorId string) (*uint64, error) {
	val, err := r.cache.Get(c, db.KEY_MIN_PURCHASED_QTY+vendorId).Result()
	if err != nil {
		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()

		tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return nil, err
		}

		defer func() {
			tx.Commit()
		}()

		rows, err := tx.QueryxContext(
			ctx,
			fmt.Sprintf(query.MinPackQtyPurchased, vendorId),
		)

		minPurchased := new(uint64)
		defer rows.Close()
		for rows.Next() {
			if err = rows.Scan(minPurchased); err != nil {
				return nil, err
			}
		}
		if minPurchased == nil {
			return nil, &core.ErrorResp{Message: "Error, min purchased qty is null"}
		}

		minPurchasedBytes, err := json.Marshal(minPurchased)
		if err != nil {
			return nil, err
		}
		if err = r.cache.Set(c, db.KEY_MIN_PURCHASED_QTY+vendorId, minPurchasedBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, err
		}
		return minPurchased, nil
	} else {
		minPurchased := new(uint64)
		if err = json.Unmarshal([]byte(val), minPurchased); err != nil {
			return nil, err
		}
		return minPurchased, nil
	}
}

func (r *AnalyticsRepoImpl) MaxPackQtyPurchased(c context.Context, vendorId string) (*uint64, error) {
	val, err := r.cache.Get(c, db.KEY_MAX_PURCHASED_QTY+vendorId).Result()
	if err != nil {
		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()

		tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return nil, err
		}

		defer func() {
			tx.Commit()
		}()

		rows, err := tx.QueryxContext(
			ctx,
			fmt.Sprintf(query.MaxPackQtyPurchased, vendorId),
		)

		maxPurchased := new(uint64)
		defer rows.Close()
		for rows.Next() {
			if err = rows.Scan(maxPurchased); err != nil {
				return nil, err
			}
		}
		if maxPurchased == nil {
			return nil, &core.ErrorResp{Message: "Error, max purchased qty is null"}
		}

		maxPurchasedBytes, err := json.Marshal(maxPurchased)
		if err != nil {
			return nil, err
		}
		if err = r.cache.Set(c, db.KEY_MAX_PURCHASED_QTY+vendorId, maxPurchasedBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, err
		}
		return maxPurchased, nil
	} else {
		maxPurchased := new(uint64)
		if err = json.Unmarshal([]byte(val), maxPurchased); err != nil {
			return nil, err
		}
		return maxPurchased, nil
	}
}

func (r *AnalyticsRepoImpl) TopCustomers(c context.Context, vendorId string) ([]*model.CustomerAnalytic, error) {
	val, err := r.cache.Get(c, db.KEY_TOP_CUSTOMERS+vendorId).Result()
	if err != nil {
		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()

		tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return nil, err
		}

		defer func() {
			tx.Commit()
		}()

		rows, err := tx.QueryxContext(
			ctx,
			fmt.Sprintf(query.TopCustomers, vendorId),
		)

		topCustomers := []*model.CustomerAnalytic{}
		defer rows.Close()
		for rows.Next() {
			customer := model.CustomerAnalytic{}
			if err = rows.StructScan(&customer); err != nil {
				return nil, err
			}
			topCustomers = append(topCustomers, &customer)
		}

		topCustomerBytes, err := json.Marshal(topCustomers)
		if err != nil {
			return nil, err
		}
		if err = r.cache.Set(c, db.KEY_TOP_CUSTOMERS+vendorId, topCustomerBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, err
		}
		return topCustomers, nil
	} else {
		topCustomers := []*model.CustomerAnalytic{}
		if err = json.Unmarshal([]byte(val), &topCustomers); err != nil {
			return nil, err
		}
		return topCustomers, nil
	}
}

func (r *AnalyticsRepoImpl) PackSales(c context.Context, vendorId string, granularity string, month int64, year int64, urlPath string) ([]*model.PackSalesAnalytic, error) {
	val, err := r.cache.Get(c, urlPath).Result()
	if err != nil {
		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()

		tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return nil, err
		}

		defer func() {
			tx.Commit()
		}()

		var queryStr string
		switch granularity {
		case "day":
			dateFilter := fmt.Sprintf("%v-%v-01", year, month)
			queryStr = fmt.Sprintf(query.PackSalesByDay, dateFilter, dateFilter, vendorId)
		case "month":
			queryStr = fmt.Sprintf(query.PackSalesByMonth, year, vendorId)
		case "year":
			queryStr = fmt.Sprintf(query.PackSalesByYear, vendorId)
		}

		rows, err := tx.QueryxContext(ctx, queryStr)
		if err != nil {
			return nil, err
		}

		packSales := []*model.PackSalesAnalytic{}
		defer rows.Close()
		for rows.Next() {
			saleAnalytic := model.PackSalesAnalytic{}
			if err = rows.StructScan(&saleAnalytic); err != nil {
				return nil, err
			}
			packSales = append(packSales, &saleAnalytic)
		}

		packSalesBytes, err := json.Marshal(packSales)
		if err != nil {
			return nil, err
		}
		if err = r.cache.Set(c, urlPath, packSalesBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, err
		}
		return packSales, nil
	} else {
		packSales := []*model.PackSalesAnalytic{}
		if err = json.Unmarshal([]byte(val), &packSales); err != nil {
			return nil, err
		}
		return packSales, nil
	}
}

func (r *AnalyticsRepoImpl) PackQtySold(c context.Context, vendorId string) ([]*model.PackQtySoldAnalytic, error) {
	val, err := r.cache.Get(c, db.KEY_PACK_QTY+vendorId).Result()
	if err != nil {
		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()

		tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return nil, err
		}

		defer func() {
			tx.Commit()
		}()

		rows, err := tx.QueryxContext(ctx, fmt.Sprintf(query.PackQtySold, vendorId))
		if err != nil {
			return nil, err
		}

		packsQtySold := []*model.PackQtySoldAnalytic{}
		defer rows.Close()
		for rows.Next() {
			packQtySold := model.PackQtySoldAnalytic{}
			if err = rows.StructScan(&packQtySold); err != nil {
				return nil, err
			}
			packsQtySold = append(packsQtySold, &packQtySold)
		}

		packsQtySoldBytes, err := json.Marshal(packsQtySold)
		if err != nil {
			return nil, err
		}
		if err = r.cache.Set(c, db.KEY_PACK_QTY+vendorId, packsQtySoldBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, err
		}
		return packsQtySold, nil
	} else {
		packsQtySold := []*model.PackQtySoldAnalytic{}
		if err = json.Unmarshal([]byte(val), &packsQtySold); err != nil {
			return nil, err
		}
		return packsQtySold, nil
	}
}
