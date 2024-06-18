package repository

import (
	"context"
	"time"
	"xo-packs/core"
	"xo-packs/db"
	"xo-packs/model"

	"github.com/redis/go-redis/v9"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type LoggingRepository interface {
	LogSignIn(context.Context, string, string) (*model.SignInLog, error)
	LogAgeAgreement(context.Context, string, string) (*model.AgeAgreementLog, error)
	LogPackPurchase(context.Context, string, string, uint64, *sqlx.Tx) (*model.PackPurchaseLog, error)
	LogTokenPurchase(context.Context, string, string, uint64, *sqlx.Tx) (*model.TokenPurchaseLog, error)
	LogUserAccountCreation(context.Context, string, string, *sqlx.Tx) (*model.UserAccountCreationLog, error)
	LogUserAccountDeletion(context.Context, string, string, *sqlx.Tx) (*model.UserAccountDeletionLog, error)
	LogDeposit(context.Context, string, string, float64, *sqlx.Tx) (*model.DepositLog, error)
	LogVendorApproval(context.Context, string, string, *sqlx.Tx) (*model.VendorApprovalLog, error)
	LogVendorRemoval(context.Context, string, string, *sqlx.Tx) (*model.VendorRemovalLog, error)
}

type LoggingRepoImpl struct {
	db    *sqlx.DB
	cache *redis.Client
}

func NewLoggingRepo(db *sqlx.DB, cache *redis.Client) LoggingRepository {
	return &LoggingRepoImpl{db: db, cache: cache}
}

func (r *LoggingRepoImpl) LogSignIn(c context.Context, uid string, ip string) (*model.SignInLog, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	now := time.Now().Format("2006-01-02 15:04:05")
	login := model.SignInLog{Uid: &uid, IpAddr: &ip, LastLogin: &now}
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Insert(db.SCHEMA_SIGN_INS).
		Columns(core.ModelColumns(login)...).
		Values(core.StructValues(login)...).
		ToSql()
	if err != nil {
		return nil, err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &login, nil
}

func (r *LoggingRepoImpl) LogAgeAgreement(c context.Context, uid string, ip string) (*model.AgeAgreementLog, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	now := time.Now().Format("2006-01-02 15:04:05")
	agreementLog := model.AgeAgreementLog{Uid: &uid, IpAddr: &ip, AgreedAt: &now}
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Insert(db.SCHEMA_AGE_AGREEMENTS).
		Columns(core.ModelColumns(agreementLog)...).
		Values(core.StructValues(agreementLog)...).
		ToSql()
	if err != nil {
		return nil, err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &agreementLog, nil
}

func (r *LoggingRepoImpl) LogPackPurchase(c context.Context, uid string, clientIp string, packId uint64, tx *sqlx.Tx) (*model.PackPurchaseLog, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	packPurchaseLog := model.PackPurchaseLog{
		Uid:         &uid,
		IpAddr:      &clientIp,
		PackId:      &packId,
		PurchasedAt: &now,
	}
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Insert(db.SCHEMA_PACK_PURCHASES).
		Columns(core.ModelColumns(packPurchaseLog)...).
		Values(core.StructValues(packPurchaseLog)...).
		ToSql()
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(c, query, args...)
	if err != nil {
		return nil, err
	}
	return &packPurchaseLog, nil
}

func (r *LoggingRepoImpl) LogTokenPurchase(c context.Context, uid string, clientIp string, tokenBundleId uint64, tx *sqlx.Tx) (*model.TokenPurchaseLog, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	tokenPurchaseLog := model.TokenPurchaseLog{
		Uid:           &uid,
		IpAddr:        &clientIp,
		TokenBundleId: &tokenBundleId,
		PurchasedAt:   &now,
	}
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Insert(db.SCHEMA_TOKEN_PURCHASES).
		Columns(core.ModelColumns(tokenPurchaseLog)...).
		Values(core.StructValues(tokenPurchaseLog)...).
		ToSql()
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(c, query, args...)
	if err != nil {
		return nil, err
	}
	return &tokenPurchaseLog, nil
}

func (r *LoggingRepoImpl) LogUserAccountCreation(c context.Context, uid string, clientIp string, tx *sqlx.Tx) (*model.UserAccountCreationLog, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	userAccountCreationLog := model.UserAccountCreationLog{
		Uid:       &uid,
		IpAddr:    &clientIp,
		CreatedAt: &now,
	}
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Insert(db.SCHEMA_USER_ACCOUNT_CREATION_LOGS).
		Columns(core.ModelColumns(userAccountCreationLog)...).
		Values(core.StructValues(userAccountCreationLog)...).
		ToSql()
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(c, query, args...)
	if err != nil {
		return nil, err
	}
	return &userAccountCreationLog, nil
}

func (r *LoggingRepoImpl) LogUserAccountDeletion(c context.Context, uid string, clientIp string, tx *sqlx.Tx) (*model.UserAccountDeletionLog, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	userAccountDeletionLog := model.UserAccountDeletionLog{
		Uid:       &uid,
		IpAddr:    &clientIp,
		DeletedAt: &now,
	}
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Insert(db.SCHEMA_USER_ACCOUNT_DELETION_LOGS).
		Columns(core.ModelColumns(userAccountDeletionLog)...).
		Values(core.StructValues(userAccountDeletionLog)...).
		ToSql()
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(c, query, args...)
	if err != nil {
		return nil, err
	}
	return &userAccountDeletionLog, nil
}

func (r *LoggingRepoImpl) LogDeposit(c context.Context, uid string, clientIp string, depositAmount float64, tx *sqlx.Tx) (*model.DepositLog, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	depositLog := model.DepositLog{
		Uid:          &uid,
		IpAddr:       &clientIp,
		DollarAmount: &depositAmount,
		DepositedAt:  &now,
	}
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Insert(db.SCHEMA_DEPOSITS).
		Columns(core.ModelColumns(depositLog)...).
		Values(core.StructValues(depositLog)...).
		ToSql()
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(c, query, args...)
	if err != nil {
		return nil, err
	}
	return &depositLog, nil
}

func (r *LoggingRepoImpl) LogVendorApproval(c context.Context, vendorId string, clientIp string, tx *sqlx.Tx) (*model.VendorApprovalLog, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	vendorApprovalLog := model.VendorApprovalLog{
		VendorId:   &vendorId,
		IpAddr:     &clientIp,
		ApprovedAt: &now,
	}
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Insert(db.SCHEMA_USER_ACCOUNT_CREATION_LOGS).
		Columns(core.ModelColumns(vendorApprovalLog)...).
		Values(core.StructValues(vendorApprovalLog)...).
		ToSql()
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(c, query, args...)
	if err != nil {
		return nil, err
	}
	return &vendorApprovalLog, nil
}

func (r *LoggingRepoImpl) LogVendorRemoval(c context.Context, vendorId string, clientIp string, tx *sqlx.Tx) (*model.VendorRemovalLog, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	vendorRemovalLog := model.VendorRemovalLog{
		VendorId:  &vendorId,
		IpAddr:    &clientIp,
		RemovedAt: &now,
	}
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := psql.
		Insert(db.SCHEMA_VENDOR_REMOVAL_LOGS).
		Columns(core.ModelColumns(vendorRemovalLog)...).
		Values(core.StructValues(vendorRemovalLog)...).
		ToSql()
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(c, query, args...)
	if err != nil {
		return nil, err
	}
	return &vendorRemovalLog, nil
}
