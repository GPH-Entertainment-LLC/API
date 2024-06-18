package repository

import (
	"context"
	"xo-packs/model"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type ShippingRepository interface {
	GetShippingInfo(context.Context, string) (*model.ShippingInfo, error)
	UploadShippingInfo(context.Context, string, *model.ShippingInfo) (*model.ShippingInfo, error)
	UpdateShippingInfo(context.Context, string) error
}

type ShippingRepoImpl struct {
	db    *sqlx.DB
	cache *redis.Client
}

func NewShippingRepo(db *sqlx.DB, cache *redis.Client) ShippingRepository {
	return &ShippingRepoImpl{db: db, cache: cache}
}

func (r *ShippingRepoImpl) GetShippingInfo(c context.Context, uid string) (*model.ShippingInfo, error) {
	return nil, nil
}

func (r *ShippingRepoImpl) UploadShippingInfo(c context.Context, uid string, shippingInfo *model.ShippingInfo) (*model.ShippingInfo, error) {
	return shippingInfo, nil
}

func (r *ShippingRepoImpl) UpdateShippingInfo(c context.Context, uid string) error {
	return nil
}
