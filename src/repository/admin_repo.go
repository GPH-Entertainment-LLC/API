package repository

import (
	"context"
	"xo-packs/core"
	"xo-packs/db"
	"xo-packs/model"

	"database/sql"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type AdminRepository interface {
	ApproveVendor(context.Context, string, string) error
	RejectVendor(context.Context, string, string) error
	RemoveVendor(context.Context, string, string) error
	AddFaq(context.Context, *model.Faq) (*model.Faq, error)
	EditFaq(context.Context, map[string]interface{}, uint64) error
	RemoveFaq(context.Context, uint64) error
	FlushCache(context.Context) error
}

type AdminRepoImpl struct {
	db    *sqlx.DB
	cache *redis.Client
}

func NewAdminRepo(db *sqlx.DB, cache *redis.Client) AdminRepository {
	return &AdminRepoImpl{db: db, cache: cache}
}

func (r *AdminRepoImpl) ApproveVendor(c context.Context, uid string, uername string) error {
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
		Set("is_vendor", true).
		Where(squirrel.Eq{"uid": uid}).
		ToSql()
	if err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	applicationApproveMap := map[string]interface{}{"approved_at": now, "status": "approved"}
	query, args, err = psql.
		Update(db.SCHEMA_VENDOR_APPLICATIONS).
		SetMap(applicationApproveMap).
		Where(squirrel.Eq{"uid": uid, "status": "pending"}).
		ToSql()
	if err != nil {
		return err
	}

	result, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		fmt.Printf("error updating vendor flag to true for %v error %v: ", uid, err)
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
	return &core.ErrorResp{
		Message: "unable to approve this user to be a vendor. Check the vendor application record and user to make sure the UIDs match",
	}
}

func (r *AdminRepoImpl) RemoveVendor(c context.Context, uid string, uername string) error {
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
		fmt.Printf("error updating vendor flag to false for %v error %v: ", uid, err)
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
	return &core.ErrorResp{Message: "unable to update user is_vendor privileges"}
}

func (r *AdminRepoImpl) RejectVendor(c context.Context, uid string, uername string) error {
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
	now := time.Now().Format("2006-01-02 15:04:05")
	applicationRejectionMap := map[string]interface{}{"rejected_at": now, "status": "rejected"}
	query, args, err := psql.
		Update(db.SCHEMA_VENDOR_APPLICATIONS).
		SetMap(applicationRejectionMap).
		Where(squirrel.Eq{"uid": uid, "status": "pending"}).
		ToSql()
	if err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		fmt.Printf("error updating vendor flag to true for %v error %v: ", uid, err)
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
	return &core.ErrorResp{
		Message: "unable to reject this user to be a vendor. Check the vendor application record and user to make sure the UIDs match",
	}
}

func (r *AdminRepoImpl) AddFaq(c context.Context, faq *model.Faq) (*model.Faq, error) {
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
	now := time.Now().Format("2006-01-02 15:04:05")
	active := true
	faq.CreatedAt = &now
	faq.Active = &active
	query, args, err := psql.
		Insert(db.SCHEMA_FAQS).
		Columns(core.ModelColumns(faq)...).
		Values(core.StructValues(faq)...).
		ToSql()
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return faq, nil
}

func (r *AdminRepoImpl) EditFaq(c context.Context, patchMap map[string]interface{}, id uint64) error {
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
		Update(db.SCHEMA_FAQS).
		SetMap(patchMap).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *AdminRepoImpl) RemoveFaq(c context.Context, id uint64) error {
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
	now := time.Now().Format("2006-01-02 15:04:05")
	query, args, err := psql.
		Update(db.SCHEMA_FAQS).
		SetMap(map[string]interface{}{"removed_at": now, "active": false}).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *AdminRepoImpl) FlushCache(c context.Context) error {
	return r.cache.FlushAll(c).Err()
}
