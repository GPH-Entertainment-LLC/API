package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"xo-packs/model"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type FinancialRepository interface {
	GetCreatorEarnings(context.Context, string) ([]*model.CreatorEarningsPeriod, error)
	GetReferralEarnings(context.Context, string) ([]*model.ReferralEarningsPeriod, error)
	GetAllEarnings(context.Context, string) ([]*model.AllEarningsPeriod, error)
}

type FinancialRepoImpl struct {
	db    *sqlx.DB
	cache *redis.Client
}

func NewFinancialRepo(db *sqlx.DB, cache *redis.Client) FinancialRepository {
	return &FinancialRepoImpl{db: db, cache: cache}
}

func (r FinancialRepoImpl) GetCreatorEarnings(c context.Context, creatorUid string) ([]*model.CreatorEarningsPeriod, error) {
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
		Select(
			"uid",
			"email",
			"starting_period",
			"ending_period",
			"payout_date",
			"earning_type",
			"earnings",
			"current_period",
			"payout_status",
		).
		From("financial.v_creator_earnings").
		Where(squirrel.Eq{"uid": creatorUid}).
		OrderBy("payout_date desc").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	creatorEarningPeriods := []*model.CreatorEarningsPeriod{}
	for rows.Next() {
		creatorEarningPeriod := model.CreatorEarningsPeriod{}
		if err := rows.StructScan(&creatorEarningPeriod); err != nil {
			return nil, err
		}
		fmt.Println("EARNINGS", *creatorEarningPeriod.Earnings)
		creatorEarningPeriods = append(creatorEarningPeriods, &creatorEarningPeriod)
	}
	return creatorEarningPeriods, nil
}

func (r FinancialRepoImpl) GetReferralEarnings(c context.Context, creatorUid string) ([]*model.ReferralEarningsPeriod, error) {
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
		Select(
			"uid",
			"email",
			"starting_period",
			"ending_period",
			"payout_date",
			"earning_type",
			"earnings",
			"current_period",
			"payout_status",
		).
		From("financial.v_referral_earnings").
		Where(squirrel.Eq{"uid": creatorUid}).
		OrderBy("payout_date desc").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	referralEarningPeriods := []*model.ReferralEarningsPeriod{}
	for rows.Next() {
		referralEarningPeriod := model.ReferralEarningsPeriod{}
		if err := rows.StructScan(&referralEarningPeriod); err != nil {
			return nil, err
		}
		referralEarningPeriods = append(referralEarningPeriods, &referralEarningPeriod)
	}
	return referralEarningPeriods, nil
}

func (r FinancialRepoImpl) GetAllEarnings(c context.Context, creatorUid string) ([]*model.AllEarningsPeriod, error) {
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
		Select(
			"uid",
			"email",
			"starting_period",
			"ending_period",
			"payout_date",
			"earning_type",
			"earnings",
			"current_period",
			"payout_status",
		).
		From("financial.v_all_earnings").
		Where(squirrel.Eq{"uid": creatorUid}).
		OrderBy("payout_date desc").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	allEarningPeriods := []*model.AllEarningsPeriod{}
	for rows.Next() {
		allEarningPeriod := model.AllEarningsPeriod{}
		if err := rows.StructScan(&allEarningPeriod); err != nil {
			return nil, err
		}
		allEarningPeriods = append(allEarningPeriods, &allEarningPeriod)
	}
	return allEarningPeriods, nil
}
