package service

import (
	"context"
	"mime/multipart"
	"xo-packs/core"
	"xo-packs/model"
	"xo-packs/repository"
)

type ApplicationService interface {
	GetPendingApplication(context.Context, string) (*model.VendorApplication, error)
	GetApprovedApplication(context.Context, string) (*model.VendorApplication, error)
	GetRejectedApplication(context.Context, string) (*model.VendorApplication, error)
	ApplicationSubmit(context.Context, *model.VendorApplication, *multipart.FileHeader, *multipart.FileHeader, *multipart.FileHeader) (*model.VendorApplication, error)
	ClearApplicationCache(context.Context, string, string) error
}

type ApplicationSvcImpl struct {
	applicationRepo repository.ApplicationRepository
}

func NewApplicationService(repo repository.ApplicationRepository) ApplicationService {
	return &ApplicationSvcImpl{applicationRepo: repo}
}

func (service *ApplicationSvcImpl) ApplicationSubmit(
	c context.Context, vendorApplication *model.VendorApplication, frontIdFile *multipart.FileHeader, backIdFile *multipart.FileHeader, profileIdFile *multipart.FileHeader) (*model.VendorApplication, error) {

	existingApplication, err := service.GetPendingApplication(c, *vendorApplication.Uid)
	if err != nil {
		return nil, err
	}

	if existingApplication.ID != nil {
		return nil, &core.ErrorResp{
			Message: "You have already submitted an application and must wait for it to be either accepted or rejected before submitting another",
		}
	}

	return service.applicationRepo.ApplicationSubmit(c, vendorApplication, frontIdFile, backIdFile, profileIdFile)
}

func (service *ApplicationSvcImpl) GetPendingApplication(c context.Context, uid string) (*model.VendorApplication, error) {
	return service.applicationRepo.GetPendingApplication(c, uid)
}

func (service *ApplicationSvcImpl) GetApprovedApplication(c context.Context, uid string) (*model.VendorApplication, error) {
	return service.applicationRepo.GetApprovedApplication(c, uid)
}

func (service *ApplicationSvcImpl) GetRejectedApplication(c context.Context, uid string) (*model.VendorApplication, error) {
	return service.applicationRepo.GetRejectedApplication(c, uid)
}

func (service *ApplicationSvcImpl) ClearApplicationCache(c context.Context, uid string, status string) error {
	return service.applicationRepo.ClearApplicationCache(c, uid, status)
}
