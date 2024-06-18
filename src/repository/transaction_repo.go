package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"
	"xo-packs/core"
	"xo-packs/db"
	"xo-packs/model"
	"xo-packs/query"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type TransactionRepository interface {
	NewSale(context.Context, *model.NewSalesTransaction) (*model.NewSalesTransaction, error)
	GetUserTransactionInfo(context.Context, string) (*model.UserTransactionInfo, error)
	ChargeTransaction(context.Context, *model.Transaction) (*model.Transaction, error)
	GetUserTransactionHistoryPage(context.Context, string, uint64) ([]*model.PageUserTransactionHistory, *uint64, error)
	GetCharge(context.Context)
}

type TransactionRepoImpl struct {
	db    *sqlx.DB
	cache *redis.Client
}

func NewTransactionRepo(db *sqlx.DB, cache *redis.Client) TransactionRepository {
	return &TransactionRepoImpl{db: db, cache: cache}
}

func (r *TransactionRepoImpl) GetUserTransactionHistoryPage(c context.Context, uid string, pageNumber uint64) ([]*model.PageUserTransactionHistory, *uint64, error) {
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

	pageSizeStr := os.Getenv("TRANSACTION_HISTORY_PAGE_SIZE")
	pageSize, err := strconv.ParseUint(pageSizeStr, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	rows, err := tx.QueryxContext(ctx, fmt.Sprintf(
		query.UserTransactionHistoryPageQuery,
		uid,
		pageSize,
		pageSize*(pageNumber)),
	)
	if err != nil {
		return nil, nil, err
	}

	defer rows.Close()
	pageTransactions := []*model.PageUserTransactionHistory{}
	for rows.Next() {
		pageTransactionHistory := model.PageUserTransactionHistory{}
		if err = rows.StructScan(&pageTransactionHistory); err != nil {
			return nil, nil, err
		}
		pageTransactions = append(pageTransactions, &pageTransactionHistory)
	}
	transactionAmount := new(uint64)
	if len(pageTransactions) <= 0 {
		*transactionAmount = 0
	} else {
		transactionAmount = pageTransactions[len(pageTransactions)-1].TransactionCount
	}

	return pageTransactions, transactionAmount, nil
}

func (r *TransactionRepoImpl) NewSale(c context.Context, txn *model.NewSalesTransaction) (*model.NewSalesTransaction, error) {
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
		Insert(db.SCHEMA_NEW_SALES_TRANSACTIONS).
		Columns(core.ModelColumns(txn)...).
		Values(core.StructValues(txn)...).
		ToSql()
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return txn, nil
}

func (r *TransactionRepoImpl) GetUserTransactionInfo(c context.Context, uid string) (*model.UserTransactionInfo, error) {
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
		Select(
			"uid",
			"token_bundle_id",
			"transaction_id",
			"subscription_id",
			"client_accnum",
			"client_subacc",
			"username",
			"password",
			"last4",
			"payment_type",
			"card_type",
			"billed_initial_price",
			"initial_period",
			"billed_recurring_price",
			"recurring_period",
			"rebills",
			"currency_code",
		).
		From(db.SCHEMA_NEW_SALES_TRANSACTIONS).
		Where(squirrel.And{squirrel.Eq{"uid": uid}, squirrel.NotEq{"transaction_id": nil}}).
		OrderBy("tran_datetime desc").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	userInfo := model.UserTransactionInfo{}
	defer rows.Close()
	for rows.Next() {
		if err = rows.StructScan(&userInfo); err != nil {
			return nil, err
		}
	}

	if userInfo.Uid == nil {
		return nil, nil
	} else {
		return &userInfo, nil
	}
}

func (r *TransactionRepoImpl) ChargeTransaction(c context.Context, txn *model.Transaction) (*model.Transaction, error) {
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
		Insert(db.SCHEMA_TRANSACTIONS).
		Columns(
			"uid",
			"token_bundle_id",
			"transaction_id",
			"subscription_id",
			"tran_datetime",
			"client_accnum",
			"client_subacc",
			"username",
			"password",
			"billed_initial_price",
			"initial_period",
			"billed_recurring_price",
			"recurring_period",
			"rebills",
			"currency_code",
		).
		Values(
			txn.Uid,
			txn.TokenBundleId,
			txn.TransactionId,
			txn.SubscriptionId,
			txn.TranDatetime,
			txn.ClientAccNum,
			txn.ClientSubAcc,
			txn.Username,
			txn.Password,
			txn.InitialPrice,
			txn.InitialPeriod,
			txn.RecurringPrice,
			txn.RecurringPeriod,
			txn.Rebills,
			txn.CurrencyCode,
		).
		ToSql()
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return txn, nil
}

func (r *TransactionRepoImpl) GetCharge(c context.Context) {
	return
}
