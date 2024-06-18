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

type TokenRepository interface {
	AddBundle(context.Context, *float64, *float64, string) (*model.TokenBundle, error)
	ActiveTokenRate(context.Context) (*model.TokenCurrencyRate, error)
	GetBundle(context.Context, uint64) (*model.TokenBundle, error)
	GetBalance(context.Context, string) (*model.TokenBalance, error)
	GetCurrentBundles(context.Context) ([]*model.TokenBundle, error)
	GetBundlesByPrice(context.Context, *float64) ([]*model.TokenBundle, error)
	GetBundleByPrice(context.Context, *float64) (*model.TokenBundle, error)
	GetBundlesByPriceRange(context.Context, *float64, *float64) ([]*model.TokenBundle, error)
	GetInactiveBundles(context.Context) ([]*model.TokenBundle, error)
	BuyTokens(context.Context, string, uint64, string) (*model.TokenOrder, error)
	UpdateBalance(context.Context, string, map[string]interface{}) (*model.TokenBalance, error)
	DeleteBundle(context.Context, uint64) (*model.TokenBundle, error)
	ClearUserTokenCache(context.Context, string) error
}

type TokenRepoImpl struct {
	db    *sqlx.DB
	cache *redis.Client
}

type TokenError struct {
	message string
}

func NewTokenRepository(db *sqlx.DB, cache *redis.Client) TokenRepository {
	tokenRepo := TokenRepoImpl{db: db, cache: cache}
	return &tokenRepo
}

func (e *TokenError) Error() string {
	return e.message
}

// TODO -> call the id service generator lambda
func (r *TokenRepoImpl) AddBundle(c context.Context, dollarAmt *float64, tokenAmt *float64, bundleImageUrl string) (*model.TokenBundle, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	createdAt := time.Now().Format("2006-01-02 15:04:05")
	active := true
	newBundle := model.TokenBundle{
		DollarAmount:   dollarAmt,
		TokenAmount:    tokenAmt,
		CreatedAt:      &createdAt,
		Active:         &active,
		BundleImageUrl: &bundleImageUrl,
	}
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Insert(db.SCHEMA_TOKEN_BUNDLE).
		Columns("dollar_amount", "token_amount", "bundle_image_id", "created_at").
		Values(newBundle.DollarAmount, newBundle.TokenAmount, newBundle.Active, newBundle.CreatedAt).
		ToSql()
	if err != nil {
		return nil, err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return &newBundle, nil
}

func (r *TokenRepoImpl) ActiveTokenRate(c context.Context) (*model.TokenCurrencyRate, error) {
	val, err := r.cache.Get(c, db.KEY_ACTIVE_TOKEN_RATE).Result()
	if err != nil {
		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()

		psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
		query, args, err := psql.
			Select(
				"id",
				"token_amount",
				"dollar_amount",
				"start_date",
				"end_date",
			).
			From("financial.token_currency_rate").
			Where(squirrel.Eq{"end_date": nil}).
			Limit(1).
			ToSql()
		if err != nil {
			return nil, err
		}

		rows, err := r.db.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}

		activeTokenRate := model.TokenCurrencyRate{}
		defer rows.Close()
		for rows.Next() {
			if err = rows.StructScan(&activeTokenRate); err != nil {
				return nil, err
			}
		}

		tokenRateBytes, err := json.Marshal(activeTokenRate)
		if err != nil {
			return nil, err
		}
		if err = r.cache.Set(c, db.KEY_ACTIVE_TOKEN_RATE, tokenRateBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, err
		}
		return &activeTokenRate, nil
	} else {
		activeTokenRate := model.TokenCurrencyRate{}
		if err = json.Unmarshal([]byte(val), &activeTokenRate); err != nil {
			return nil, err
		}
		return &activeTokenRate, nil
	}
}

func (r *TokenRepoImpl) BuyTokens(c context.Context, uid string, tokenBundleId uint64, transactionId string) (*model.TokenOrder, error) {
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

	// get current active token rate
	activeTokenRate, err := r.ActiveTokenRate(ctx)
	if err != nil {
		return nil, err
	}

	// get the token bundle
	tokenBundle, err := r.GetBundle(ctx, tokenBundleId)
	if err != nil {
		return nil, err
	}

	if tokenBundle.TokenAmount == nil {
		return nil, &core.ErrorResp{
			Message: "ERROR: token bundle is null and does not have a valid token amount",
		}
	}

	// create the token order
	now := time.Now().Format("2006-01-02 15:04:05")
	tokenOrder := model.TokenOrder{
		Uid:           &uid,
		TransactionId: &transactionId,
		TokenBundleId: &tokenBundleId,
		PriceUsd:      tokenBundle.DollarAmount,
		TokenRateId:   activeTokenRate.ID,
		OrderedAt:     &now,
	}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Insert(db.SCHEMA_TOKEN_ORDERS).
		Columns(core.ModelColumns(tokenOrder)...).
		Values(core.StructValues(tokenOrder)...).
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
	if rowsAffected <= 0 {
		return nil, &core.ErrorResp{
			Message: "an error occurred placing the token order",
		}
	}

	// update the user token balance
	query, args, err = psql.
		Update(db.SCHEMA_TOKEN_BALANCE).
		Set("balance", squirrel.Expr("balance + ?", *tokenBundle.TokenAmount)).
		Where(squirrel.Eq{"uid": uid}).
		ToSql()

	result, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected <= 0 {
		return nil, &core.ErrorResp{
			Message: "an error occurred placing the token order",
		}
	}

	// commit transaction
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &tokenOrder, nil
}

func (r *TokenRepoImpl) GetBundle(c context.Context, id uint64) (*model.TokenBundle, error) {
	val, err := r.cache.Get(c, db.KEY_TOKEN_BUNDLE+fmt.Sprintf("%v", id)).Result()
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

		tokenBundle := model.TokenBundle{}
		psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
		query, args, err := psql.
			Select(
				"id",
				"dollar_amount",
				"token_amount",
				"created_at",
				"deleted_at",
				"bundle_image_url",
				"active",
			).
			From(db.SCHEMA_TOKEN_BUNDLE).
			Where(squirrel.Eq{"id": id}).
			ToSql()
		if err != nil {
			return nil, err
		}

		rows, err := tx.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}

		defer rows.Close()
		for rows.Next() {
			if err := rows.StructScan(&tokenBundle); err != nil {
				return nil, err
			}
		}

		tokenBundleBytes, err := json.Marshal(tokenBundle)
		if err != nil {
			return nil, err
		}
		if err = r.cache.Set(c, db.KEY_TOKEN_BUNDLE+fmt.Sprintf("%v", id), tokenBundleBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, err
		}
		return &tokenBundle, nil
	} else {
		tokenBundle := model.TokenBundle{}
		if err = json.Unmarshal([]byte(val), &tokenBundle); err != nil {
			return nil, err
		}
		return &tokenBundle, nil
	}
}

func (r *TokenRepoImpl) GetBalance(c context.Context, uid string) (*model.TokenBalance, error) {
	val, err := r.cache.Get(c, db.KEY_TOKEN_BALANCE+uid).Result()
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
			Select("id", "uid", "balance", "updated_at").
			From(db.SCHEMA_TOKEN_BALANCE).
			Where(squirrel.Eq{"uid": uid}).
			ToSql()
		if err != nil {
			return nil, err
		}

		rows, err := tx.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, &TokenError{message: fmt.Sprintf("An error occured reading the balance record for uid: %v", uid)}
		}

		tokenBalance := model.TokenBalance{}
		defer rows.Close()
		for rows.Next() {
			if err := rows.StructScan(&tokenBalance); err != nil {
				return nil, err
			}
		}

		if tokenBalance.ID == nil {
			return nil, &TokenError{message: fmt.Sprintf(
				"No balance record exists for uid: %v. This indicates the user does not exist or the initial balance record trigger malfunctioned", uid,
			)}
		}

		tokenBalanceBytes, err := json.Marshal(tokenBalance)
		if err != nil {
			return nil, err
		}

		if err = r.cache.Set(c, db.KEY_TOKEN_BALANCE+uid, tokenBalanceBytes, time.Duration(time.Second*3600)).Err(); err != nil {
			return nil, err
		}
		return &tokenBalance, nil
	} else {
		tokenBalance := model.TokenBalance{}
		if err = json.Unmarshal([]byte(val), &tokenBalance); err != nil {
			return nil, err
		}
		return &tokenBalance, nil
	}
}

func (r *TokenRepoImpl) GetCurrentBundles(c context.Context) ([]*model.TokenBundle, error) {
	val, err := r.cache.Get(c, db.KEY_TOKEN_BUNDLES).Result()
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

		tokenBundles := []*model.TokenBundle{}
		psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
		query, args, err := psql.
			Select("*").
			From(db.SCHEMA_TOKEN_BUNDLE).
			Where(squirrel.Eq{"active": true}).
			OrderBy("id asc").
			ToSql()
		if err != nil {
			return nil, err
		}

		rows, err := tx.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}

		defer rows.Close()
		for rows.Next() {
			currBundle := model.TokenBundle{}
			if err := rows.StructScan(&currBundle); err != nil {
				return nil, err
			}
			tokenBundles = append(tokenBundles, &currBundle)
		}

		tokenBundlesBytes, err := json.Marshal(tokenBundles)
		if err != nil {
			return nil, err
		}
		if err = r.cache.Set(c, db.KEY_TOKEN_BUNDLES, tokenBundlesBytes, time.Duration(3600)*time.Second).Err(); err != nil {
			return nil, err
		}
		return tokenBundles, nil
	} else {
		tokenBundles := []*model.TokenBundle{}
		if err = json.Unmarshal([]byte(val), &tokenBundles); err != nil {
			return nil, err
		}
		return tokenBundles, nil
	}
}

func (r *TokenRepoImpl) GetBundlesByPrice(c context.Context, dollarAmt *float64) ([]*model.TokenBundle, error) {
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

	tokenBundles := []*model.TokenBundle{}
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Select("*").
		From(db.SCHEMA_TOKEN_BUNDLE).
		Where(squirrel.Eq{"dollar_amount": *dollarAmt}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		currBundle := model.TokenBundle{}
		if err := rows.StructScan(&currBundle); err != nil {
			return nil, err
		}
		tokenBundles = append(tokenBundles, &currBundle)
	}
	return tokenBundles, nil
}

func (r *TokenRepoImpl) GetBundleByPrice(c context.Context, dollarAmt *float64) (*model.TokenBundle, error) {
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
		Select("*").
		From(db.SCHEMA_TOKEN_BUNDLE).
		Where(squirrel.Eq{"dollar_amount": *dollarAmt, "deleted_at": nil}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	currBundle := model.TokenBundle{}
	for rows.Next() {
		if err := rows.StructScan(&currBundle); err != nil {
			return nil, err
		}
	}
	return &currBundle, nil
}

func (r *TokenRepoImpl) GetBundlesByPriceRange(c context.Context, lowerDollarAmt *float64, upperDollarAmt *float64) ([]*model.TokenBundle, error) {
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

	tokenBundles := []*model.TokenBundle{}
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Select("*").
		From(db.SCHEMA_TOKEN_BUNDLE).
		Where(squirrel.And{squirrel.GtOrEq{"dollar_amount": *lowerDollarAmt}, squirrel.LtOrEq{"dollar_amount": *upperDollarAmt}}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		currBundle := model.TokenBundle{}
		if err := rows.StructScan(&currBundle); err != nil {
			return nil, err
		}
		tokenBundles = append(tokenBundles, &currBundle)
	}
	return tokenBundles, nil
}

func (r *TokenRepoImpl) GetInactiveBundles(c context.Context) ([]*model.TokenBundle, error) {
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

	tokenBundles := []*model.TokenBundle{}
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Select("*").
		From(db.SCHEMA_TOKEN_BUNDLE).
		Where(squirrel.NotEq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var currBundle model.TokenBundle
		if err := rows.StructScan(&currBundle); err != nil {
			return nil, err
		}
		tokenBundles = append(tokenBundles, &currBundle)
	}
	return tokenBundles, nil
}

func (r *TokenRepoImpl) UpdateBalance(c context.Context, uid string, patchMap map[string]interface{}) (*model.TokenBalance, error) {
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

	patchMap["updated_at"] = time.Now().Format("2006-01-02 15:04:05")
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Update(db.SCHEMA_TOKEN_BALANCE).
		SetMap(patchMap).
		Where(squirrel.Eq{"uid": uid}).
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
		// getting new updated token balance
		tokenBalance, err := r.GetBalance(c, uid)
		if err != nil {
			return nil, err
		}
		return tokenBalance, nil
	}
	err = &TokenError{message: fmt.Sprintf("Unable to update the balance for uid: %v", uid)}
	return nil, err
}

func (r *TokenRepoImpl) DeleteBundle(c context.Context, id uint64) (*model.TokenBundle, error) {
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

	deletedAt := time.Now().Format("2006-01-02 15:04:05")
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Update(db.SCHEMA_TOKEN_BUNDLE).
		Set("deleted_at", deletedAt).
		Where(squirrel.Eq{"id": id}).
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
		// remove cached bundles
		keysToDelete := []string{db.KEY_TOKEN_BUNDLES, db.KEY_TOKEN_BUNDLE + fmt.Sprintf("%v", id)}
		if err = r.cache.Del(c, keysToDelete...).Err(); err != nil {
			return nil, err
		}

		tokenBundle, err := r.GetBundle(c, id)
		if err != nil {
			return nil, err
		}

		return tokenBundle, nil
	}

	err = &TokenError{message: fmt.Sprintf("Token bundle with id: %v does not exist", id)}
	return nil, err
}

func (r *TokenRepoImpl) ClearUserTokenCache(c context.Context, uid string) error {
	return r.cache.Del(c, db.KEY_TOKEN_BALANCE+uid).Err()
}
