package core

import (
	"context"
	"xo-packs/model"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

func GetFaqs(ctx context.Context, db *sqlx.DB) ([]*model.Faq, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Select("id", "question", "answer", "active", "created_at", "removed_at").
		From("main.faqs").
		Where(squirrel.Eq{"active": true}).
		OrderBy("id asc").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	faqs := []*model.Faq{}
	for rows.Next() {
		faq := model.Faq{}
		if err := rows.StructScan(&faq); err != nil {
			return nil, err
		}
		faqs = append(faqs, &faq)
	}
	return faqs, nil
}