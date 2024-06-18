package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"os"
	"time"
	"xo-packs/core"
	"xo-packs/db"
	"xo-packs/model"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type ApplicationRepository interface {
	ApplicationSubmit(context.Context, *model.VendorApplication, *multipart.FileHeader, *multipart.FileHeader, *multipart.FileHeader) (*model.VendorApplication, error)
	GetPendingApplication(context.Context, string) (*model.VendorApplication, error)
	GetApprovedApplication(context.Context, string) (*model.VendorApplication, error)
	GetRejectedApplication(context.Context, string) (*model.VendorApplication, error)
	ClearApplicationCache(context.Context, string, string) error
}

type ApplicationRepoImpl struct {
	db    *sqlx.DB
	cache *redis.Client
}

func NewApplicationRepo(db *sqlx.DB, cache *redis.Client) ApplicationRepository {
	return &ApplicationRepoImpl{db: db, cache: cache}
}

func (r *ApplicationRepoImpl) ApplicationSubmit(
	c context.Context, vendorApplication *model.VendorApplication, frontIdFile *multipart.FileHeader, backIdFile *multipart.FileHeader, profileIdFile *multipart.FileHeader) (*model.VendorApplication, error) {

	bucket := os.Getenv("CREATOR_ID_BUCKET")

	fmt.Println("Inside repo layer")

	// upload front to S3
	key := "front/" + *vendorApplication.Uid
	if err := core.UploadImage(frontIdFile, key, bucket); err != nil {
		return nil, err
	}

	// upload back to S3
	key = "back/" + *vendorApplication.Uid
	if err := core.UploadImage(backIdFile, key, bucket); err != nil {
		return nil, err
	}

	// upload profile with Id to S3
	key = "profile/" + *vendorApplication.Uid
	if err := core.UploadImage(profileIdFile, key, bucket); err != nil {
		return nil, err
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
	vendorApplication.SubmittedAt = &now
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Insert(db.SCHEMA_VENDOR_APPLICATIONS).
		Columns(core.ModelColumns(vendorApplication)...).
		Values(core.StructValues(vendorApplication)...).
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
	return vendorApplication, nil
}

func (r *ApplicationRepoImpl) GetPendingApplication(c context.Context, uid string) (*model.VendorApplication, error) {
	val, err := r.cache.Get(c, db.KEY_PENDING_APPLICATION+uid).Result()
	if err != nil {
		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()

		psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
		query, args, err := psql.
			Select(core.VendorApplicationFieldList...).
			From(db.SCHEMA_VENDOR_APPLICATIONS).
			Where(squirrel.Eq{"uid": uid, "status": "pending"}).
			ToSql()
		if err != nil {
			return nil, err
		}

		rows, err := r.db.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}

		vendorApplication := model.VendorApplication{}
		for rows.Next() {
			if err = rows.StructScan(&vendorApplication); err != nil {
				return nil, err
			}
		}

		vendorApplicationBytes, err := json.Marshal(vendorApplication)
		if err != nil {
			return nil, err
		}

		if err = r.cache.Set(c, db.KEY_PENDING_APPLICATION, vendorApplicationBytes, time.Duration(time.Second*3600)).Err(); err != nil {
			return nil, err
		}

		return &vendorApplication, nil
	} else {
		vendorApplication := model.VendorApplication{}
		if err = json.Unmarshal([]byte(val), &vendorApplication); err != nil {
			return nil, err
		}
		return &vendorApplication, nil
	}
}

func (r *ApplicationRepoImpl) GetApprovedApplication(c context.Context, uid string) (*model.VendorApplication, error) {
	val, err := r.cache.Get(c, db.KEY_APPROVED_APPLICATION+uid).Result()
	if err != nil {
		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()

		psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
		query, args, err := psql.
			Select(core.VendorApplicationFieldList...).
			From(db.SCHEMA_VENDOR_APPLICATIONS).
			Where(squirrel.Eq{"uid": uid, "status": "approved"}).
			ToSql()
		if err != nil {
			return nil, err
		}

		rows, err := r.db.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}

		vendorApplication := model.VendorApplication{}
		for rows.Next() {
			if err = rows.StructScan(&vendorApplication); err != nil {
				return nil, err
			}
		}

		vendorApplicationBytes, err := json.Marshal(vendorApplication)
		if err != nil {
			return nil, err
		}

		if err = r.cache.Set(c, db.KEY_APPROVED_APPLICATION, vendorApplicationBytes, time.Duration(time.Second*3600)).Err(); err != nil {
			return nil, err
		}

		return &vendorApplication, nil
	} else {
		vendorApplication := model.VendorApplication{}
		if err = json.Unmarshal([]byte(val), &vendorApplication); err != nil {
			return nil, err
		}
		return &vendorApplication, nil
	}
}

func (r *ApplicationRepoImpl) GetRejectedApplication(c context.Context, uid string) (*model.VendorApplication, error) {
	val, err := r.cache.Get(c, db.KEY_REJECTED_APPLICATION+uid).Result()
	if err != nil {
		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()

		psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
		query, args, err := psql.
			Select(core.VendorApplicationFieldList...).
			From(db.SCHEMA_VENDOR_APPLICATIONS).
			Where(squirrel.Eq{"uid": uid, "status": "rejected"}).
			ToSql()
		if err != nil {
			return nil, err
		}

		rows, err := r.db.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}

		vendorApplication := model.VendorApplication{}
		for rows.Next() {
			if err = rows.StructScan(&vendorApplication); err != nil {
				return nil, err
			}
		}

		vendorApplicationBytes, err := json.Marshal(vendorApplication)
		if err != nil {
			return nil, err
		}

		if err = r.cache.Set(c, db.KEY_REJECTED_APPLICATION, vendorApplicationBytes, time.Duration(time.Second*3600)).Err(); err != nil {
			return nil, err
		}

		return &vendorApplication, nil
	} else {
		vendorApplication := model.VendorApplication{}
		if err = json.Unmarshal([]byte(val), &vendorApplication); err != nil {
			return nil, err
		}
		return &vendorApplication, nil
	}
}

func (r *ApplicationRepoImpl) ClearApplicationCache(c context.Context, uid string, status string) error {
	cacheKey := ""
	switch status {
	case "pending":
		cacheKey = db.KEY_PENDING_APPLICATION
	case "approved":
		cacheKey = db.KEY_APPROVED_APPLICATION
	case "rejected":
		cacheKey = db.KEY_REJECTED_APPLICATION
	}
	return r.cache.Del(c, cacheKey+uid).Err()
}
