package core

import (
	"context"
	"fmt"
	"xo-packs/db"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

func GetSortList(category string) ([]string, error) {
	if category == "vendors" {
		return []string{"created_at", "first_name", "last_name", "display_name"}, nil
	} else if category == "packs" {
		return []string{"vendor_id", "created_at", "latest_purchased_at", "token_amount", "pack_amount", "item_qty"}, nil
	} else if category == "items" {
		return []string{"latest_acquired_at", "vendor_id", "rarity", "name"}, nil
	}
	return nil, &ErrorResp{Message: "category must be vendors, packs, items"}
}

func GetCategories(ctx context.Context, table string, tx *sqlx.Tx, err error) ([]string, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Select("distinct category").
		From("main.categories").
		Join(fmt.Sprintf("main.%v c on categories.id = c.category_id", table)).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := tx.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	categories := []string{}
	for rows.Next() {
		category := ""
		if err := rows.Scan(&category); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func GetCategoryIds(ctx context.Context, categories []string, tx *sqlx.Tx, err error) ([]uint64, error) {
	categoryIds := []uint64{}
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Select("id").
		From(db.SCHEMA_CATEGORIES).
		Where(squirrel.Eq{"category": categories}).
		ToSql()
	if err != nil {
		return categoryIds, err
	}

	rows, err := tx.QueryxContext(ctx, query, args...)
	defer rows.Close()
	for rows.Next() {
		categoryId := uint64(0)
		if err := rows.Scan(&categoryId); err != nil {
			return categoryIds, err
		}
		categoryIds = append(categoryIds, categoryId)
	}
	fmt.Println("category ids: ", categoryIds)
	return categoryIds, nil
}
