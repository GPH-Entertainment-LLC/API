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

type ReportRepository interface {
	GetReportOpts(context.Context) ([]*model.ReportOpt, error)
	SubmitReport(context.Context, *model.Report) (*model.ReportExpanded, error)
	GetSubmittedReport(context.Context, string, string) (*model.Report, error)
	GetExpandedReport(context.Context, uint64) (*model.ReportExpanded, error)
}

type ReportRepoImpl struct {
	db    *sqlx.DB
	cache *redis.Client
}

func NewReportRepo(db *sqlx.DB, cache *redis.Client) ReportRepository {
	return &ReportRepoImpl{db: db, cache: cache}
}

func (r *ReportRepoImpl) GetReportOpts(c context.Context) ([]*model.ReportOpt, error) {
	val, err := r.cache.Get(c, db.KEY_REPORT_OPTS).Result()
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
			Select(core.ReportOptFieldList...).
			From(db.SCHEMA_REPORT_OPTS).
			Where(squirrel.Eq{"active": true}).
			ToSql()
		if err != nil {
			return nil, err
		}

		rows, err := tx.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}

		reportOpts := []*model.ReportOpt{}
		for rows.Next() {
			opt := model.ReportOpt{}
			if err = rows.StructScan(&opt); err != nil {
				return nil, err
			}
			reportOpts = append(reportOpts, &opt)
		}

		reportOptBytes, err := json.Marshal(reportOpts)
		if err != nil {
			return nil, err
		}

		if err = r.cache.Set(c, db.KEY_REPORT_OPTS, reportOptBytes, time.Duration(time.Second*3600)).Err(); err != nil {
			return nil, err
		}
		return reportOpts, err
	} else {
		reportOpts := []*model.ReportOpt{}
		if err = json.Unmarshal([]byte(val), &reportOpts); err != nil {
			return nil, err
		}
		return reportOpts, nil
	}
}

func (r *ReportRepoImpl) GetSubmittedReport(c context.Context, uid string, reportedUid string) (*model.Report, error) {
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
		Select(core.ReportFieldList...).
		From(db.SCHEMA_REPORTS).
		Where(squirrel.Eq{"uid": uid, "reported_uid": reportedUid, "active": true}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	report := model.Report{}
	for rows.Next() {
		if err = rows.StructScan(&report); err != nil {
			return nil, err
		}
	}
	return &report, err
}

func (r *ReportRepoImpl) GetExpandedReport(c context.Context, id uint64) (*model.ReportExpanded, error) {
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
		Select([]string{
			"r.id as report_id",
			"r.uid",
			"r.reported_uid",
			"r.reported_at",
			"r.notes",
			"r.active",
			"r.resolved_at",
			"opt.opt",
		}...).
		From("main.reports r").
		Join("main.report_opts opt on r.opt_id = opt.id and opt.active = true").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	reportExpanded := model.ReportExpanded{}
	for rows.Next() {
		if err = rows.StructScan(&reportExpanded); err != nil {
			return nil, err
		}
	}
	return &reportExpanded, nil
}

func (r *ReportRepoImpl) SubmitReport(c context.Context, report *model.Report) (*model.ReportExpanded, error) {
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
	report.ReportedAt = &now
	report.Active = &active
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Insert(db.SCHEMA_REPORTS).
		Columns(core.ModelColumns(report)...).
		Values(core.StructValues(report)...).
		Suffix("RETURNING \"id\"").
		ToSql()

	if err != nil {
		return nil, err
	}

	insertedId := new(uint64)
	err = tx.QueryRowContext(ctx, query, args...).Scan(insertedId)
	if err != nil {
		return nil, err
	}
	if insertedId == nil {
		return nil, &core.ErrorResp{Message: "critical error; unable to add item"}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	reportExpanded, err := r.GetExpandedReport(c, *insertedId)
	if err != nil {
		return nil, err
	}
	return reportExpanded, nil
}
