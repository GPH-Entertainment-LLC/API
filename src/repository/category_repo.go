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

	"github.com/redis/go-redis/v9"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type CategoryRepository interface {
	GetAllCategories(context.Context) ([]*model.Category, error)
	GetCategories(context.Context, string) ([]*model.Category, error)
	GetCategoryLiterals(context.Context) ([]string, error)
}

type CategoryRepoImpl struct {
	db    *sqlx.DB
	cache *redis.Client
}

func NewCategoryRepo(db *sqlx.DB, cache *redis.Client) CategoryRepository {
	return &CategoryRepoImpl{db: db, cache: cache}
}

func (r *CategoryRepoImpl) GetAllCategories(c context.Context) ([]*model.Category, error) {
	val, err := r.cache.Get(c, db.KEY_ALL_CATEGORY).Result()
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

		psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
		query, args, err := psql.
			Select("id", "category").
			From(db.SCHEMA_CATEGORIES).
			ToSql()
		if err != nil {
			return nil, err
		}

		categories := []*model.Category{}
		rows, err := tx.QueryxContext(ctx, query, args...)
		defer rows.Close()
		for rows.Next() {
			category := model.Category{}
			if err := rows.StructScan(&category); err != nil {
				return nil, err
			}
			categories = append(categories, &category)
		}

		categoryBytes, err := json.Marshal(categories)
		if err != nil {
			return nil, err
		}
		if err = r.cache.Set(c, db.KEY_ALL_CATEGORY, categoryBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, err
		}
		return categories, nil
	} else {
		categories := []*model.Category{}
		if err = json.Unmarshal([]byte(val), &categories); err != nil {
			return nil, err
		}
		return categories, nil
	}
}

func (r *CategoryRepoImpl) GetCategories(c context.Context, category string) ([]*model.Category, error) {
	switch category {
	default:
		category = ""
	case "item":
		category = "item_categories"
	case "pack":
		category = "pack_categories"
	case "vendor":
		category = "vendor_categories"
	}

	if category == "" {
		return nil, &core.ErrorResp{Message: "category must be item, pack, or vendor"}
	}
	val, err := r.cache.Get(c, db.KEY_CATEGORY+category).Result()
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

		psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
		query, args, err := psql.
			Select("c.id", "c.category").
			From("main.categories").
			Join(fmt.Sprintf("main.%v c on categories.id = c.category_id", category)).
			ToSql()
		if err != nil {
			return nil, err
		}
		rows, err := tx.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}

		categories := []*model.Category{}
		defer rows.Close()
		for rows.Next() {
			category := model.Category{}
			if err := rows.Scan(&category); err != nil {
				return nil, err
			}
			categories = append(categories, &category)
		}
		categoryBytes, err := json.Marshal(categories)
		if err != nil {
			return nil, err
		}
		if err = r.cache.Set(c, db.KEY_CATEGORY+category, categoryBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, err
		}
		return categories, nil
	} else {
		categories := []*model.Category{}
		if err = json.Unmarshal([]byte(val), &categories); err != nil {
			return nil, err
		}
		return categories, nil
	}
}

func (r *CategoryRepoImpl) GetCategoryLiterals(c context.Context) ([]string, error) {
	val, err := r.cache.Get(c, db.KEY_ALL_CATEGORY_LITERALS).Result()
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

		psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
		query, args, err := psql.
			Select("category").
			From(db.SCHEMA_CATEGORIES).
			ToSql()
		if err != nil {
			return nil, err
		}

		categories := []string{}
		rows, err := tx.QueryContext(ctx, query, args...)
		defer rows.Close()
		for rows.Next() {
			category := ""
			if err := rows.Scan(&category); err != nil {
				return nil, err
			}
			categories = append(categories, category)
		}

		categoryBytes, err := json.Marshal(categories)
		if err != nil {
			return nil, err
		}
		if err = r.cache.Set(c, db.KEY_ALL_CATEGORY_LITERALS, categoryBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, err
		}
		return categories, nil
	} else {
		categories := []string{}
		if err = json.Unmarshal([]byte(val), &categories); err != nil {
			return nil, err
		}
		return categories, nil
	}
}
