package service

import (
	"context"
	"xo-packs/model"
	"xo-packs/repository"
)

type ShippingService interface {
	GetShippingInfo(context.Context, string) (*model.ShippingInfo, error)
	UploadShippingInfo(context.Context, string, *model.ShippingInfo) (*model.ShippingInfo, error)
	UpdateShippingInfo(context.Context, string) error
}

type ShippingSvcImpl struct {
	shippingRepo repository.ShippingRepository
}

func NewShippingService(shippingRepo repository.ShippingRepository) ShippingService {
	return &ShippingSvcImpl{shippingRepo: shippingRepo}
}

func (service *ShippingSvcImpl) GetShippingInfo(c context.Context, uid string) (*model.ShippingInfo, error) {
	return service.shippingRepo.GetShippingInfo(c, uid)
}

func (service *ShippingSvcImpl) UploadShippingInfo(c context.Context, uid string, shippingInfo *model.ShippingInfo) (*model.ShippingInfo, error) {
	return service.shippingRepo.UploadShippingInfo(c, uid, shippingInfo)
}

func (service *ShippingSvcImpl) UpdateShippingInfo(c context.Context, uid string) error {
	return service.shippingRepo.UpdateShippingInfo(c, uid)
}
