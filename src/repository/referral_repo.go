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

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type ReferralRepository interface {
	CreateCode(context.Context, string, string) (*model.ReferralCode, error)
	RemoveCode(context.Context, string, string) error
	ValidateCode(context.Context, string, string, string) (*model.Referral, error)
	GetActiveCodes(context.Context, string) ([]*model.ReferralCode, error)
	GetReferralCode(context.Context, string) (*model.ReferralCode, error)
	ClearReferralCache(context.Context, string) error
	ClearReferralCodeCache(context.Context, string) error
}

type ReferralRepoImpl struct {
	db    *sqlx.DB
	cache *redis.Client
}

func NewReferralRepo(db *sqlx.DB, cache *redis.Client) ReferralRepository {
	return &ReferralRepoImpl{db: db, cache: cache}
}

func (r *ReferralRepoImpl) CreateCode(c context.Context, uid string, code string) (*model.ReferralCode, error) {
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
	active := true
	referralCode := model.ReferralCode{
		Code:      &code,
		Uid:       &uid,
		Active:    &active,
		CreatedAt: &now,
	}
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Insert(db.SCHEMA_REFERRAL_CODES).
		Columns(core.ModelColumns(referralCode)...).
		Values(core.StructValues(referralCode)...).
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
		if err = tx.Commit(); err != nil {
			return nil, err
		}
		return &referralCode, nil
	}

	return nil, &core.ErrorResp{Message: "An error occurred inserting new referral code"}
}

func (r *ReferralRepoImpl) RemoveCode(c context.Context, uid string, code string) error {
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
		Delete(db.SCHEMA_REFERRAL_CODES).
		Where(squirrel.Eq{"code": code, "uid": uid}).
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
	}
	return &core.ErrorResp{Message: "An error occurred removing the referral code. Code does not exist"}
}

func (r *ReferralRepoImpl) ValidateCode(c context.Context, referrerUid string, refereeUid string, code string) (*model.Referral, error) {
	if referrerUid == refereeUid {
		return nil, &core.ErrorResp{
			Message: "You cannot use your own referral code.",
		}
	}
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
	referral := model.Referral{
		RefereeUid:  &refereeUid,
		ReferrerUid: &referrerUid,
		Code:        &code,
		ValidatedAt: &now,
	}
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Insert(db.SCHEMA_REFERRALS).
		Columns(core.ModelColumns(referral)...).
		Values(core.StructValues(referral)...).
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
		if err = tx.Commit(); err != nil {
			return nil, err
		}
		return &referral, nil
	}
	return nil, &core.ErrorResp{Message: fmt.Sprintf("An error validating the code: %v for user: %v", code, referrerUid)}
}

func (r *ReferralRepoImpl) GetActiveCodes(c context.Context, uid string) ([]*model.ReferralCode, error) {
	val, err := r.cache.Get(c, db.KEY_ACTIVE_USER_REFERRAL_CODES+uid).Result()
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
			Select(core.ReferralCodeFieldList...).
			From(db.SCHEMA_REFERRAL_CODES).
			Where(squirrel.Eq{"uid": uid, "active": true}).
			ToSql()
		if err != nil {
			return nil, err
		}

		rows, err := tx.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}

		referralCodes := []*model.ReferralCode{}
		for rows.Next() {
			referralCode := model.ReferralCode{}
			if err = rows.StructScan(&referralCode); err != nil {
				return nil, err
			}
			referralCodes = append(referralCodes, &referralCode)
		}

		referralCodeBytes, err := json.Marshal(referralCodes)
		if err != nil {
			return nil, err
		}

		if err = r.cache.Set(c, db.KEY_ACTIVE_USER_REFERRAL_CODES+uid, referralCodeBytes, time.Duration(time.Second*3600)).Err(); err != nil {
			return nil, err
		}

		return referralCodes, nil
	} else {
		referralCodes := []*model.ReferralCode{}
		if err = json.Unmarshal([]byte(val), &referralCodes); err != nil {
			return nil, err
		}
		return referralCodes, nil
	}
}

func (r *ReferralRepoImpl) GetReferralCode(c context.Context, code string) (*model.ReferralCode, error) {
	val, err := r.cache.Get(c, db.KEY_ACTIVE_REFERRAL_CODES+code).Result()
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
			Select(core.ReferralCodeFieldList...).
			From(db.SCHEMA_REFERRAL_CODES).
			Where(squirrel.Eq{"code": code, "active": true}).
			ToSql()
		if err != nil {
			return nil, err
		}

		rows, err := tx.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}

		referralCode := model.ReferralCode{}
		for rows.Next() {
			if err = rows.StructScan(&referralCode); err != nil {
				return nil, err
			}
		}

		referralCodeBytes, err := json.Marshal(referralCode)
		if err != nil {
			return nil, err
		}

		if err = r.cache.Set(c, db.KEY_ACTIVE_REFERRAL_CODES+code, referralCodeBytes, time.Duration(time.Second*3600)).Err(); err != nil {
			return nil, err
		}

		return &referralCode, nil
	} else {
		referralCode := model.ReferralCode{}
		if err = json.Unmarshal([]byte(val), &referralCode); err != nil {
			return nil, err
		}
		return &referralCode, nil
	}
}

func (r *ReferralRepoImpl) ClearReferralCache(c context.Context, uid string) error {
	return r.cache.Del(c, db.KEY_ACTIVE_USER_REFERRAL_CODES+uid).Err()
}

func (r *ReferralRepoImpl) ClearReferralCodeCache(c context.Context, code string) error {
	return r.cache.Del(c, db.KEY_ACTIVE_REFERRAL_CODES+code).Err()
}
