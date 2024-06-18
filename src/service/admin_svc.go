package service

import (
	"context"
	"xo-packs/model"
	"xo-packs/repository"
)

type AdminService interface {
	ApproveVendor(context.Context, string, string, ApplicationService, UserService) (*model.VendorApplication, error)
	RejectVendor(context.Context, string, string, ApplicationService) (*model.VendorApplication, error)
	RemoveVendor(context.Context, string, string, UserService) error
	AddFaq(context.Context, *model.Faq) (*model.Faq, error)
	EditFaq(context.Context, map[string]interface{}, uint64) error
	RemoveFaq(context.Context, uint64) error
	FlushCache(context.Context) error
}

type AdminSvcImpl struct {
	adminRepo repository.AdminRepository
}

func NewAdminService(repo repository.AdminRepository) AdminService {
	return &AdminSvcImpl{adminRepo: repo}
}

func (service *AdminSvcImpl) ApproveVendor(c context.Context, uid string, username string, applicationService ApplicationService, userService UserService) (*model.VendorApplication, error) {
	if err := service.adminRepo.ApproveVendor(c, uid, username); err != nil {
		return nil, err
	}

	// flushing cache
	if err := service.FlushCache(c); err != nil {
		return nil, err
	}

	creatorApplication, err := applicationService.GetApprovedApplication(c, uid)
	if err != nil {
		return nil, err
	}

	return creatorApplication, nil
}

func (service *AdminSvcImpl) RejectVendor(c context.Context, uid string, username string, applicationService ApplicationService) (*model.VendorApplication, error) {
	if err := service.adminRepo.RejectVendor(c, uid, username); err != nil {
		return nil, err
	}

	if err := applicationService.ClearApplicationCache(c, uid, "rejected"); err != nil {
		return nil, err
	}

	creatorApplication, err := applicationService.GetRejectedApplication(c, uid)
	if err != nil {
		return nil, err
	}

	return creatorApplication, nil
}

func (service *AdminSvcImpl) RemoveVendor(c context.Context, uid string, username string, userService UserService) error {
	err := service.adminRepo.RemoveVendor(c, uid, username)
	if err != nil {
		return err
	}

	// flushing cache
	return service.FlushCache(c)
}

func (service *AdminSvcImpl) AddFaq(c context.Context, faq *model.Faq) (*model.Faq, error) {
	return service.adminRepo.AddFaq(c, faq)
}

func (service *AdminSvcImpl) RemoveFaq(c context.Context, id uint64) error {
	return service.adminRepo.RemoveFaq(c, id)
}

func (service *AdminSvcImpl) EditFaq(c context.Context, patchMap map[string]interface{}, id uint64) error {
	return service.adminRepo.EditFaq(c, patchMap, id)
}

func (service *AdminSvcImpl) FlushCache(c context.Context) error {
	return service.adminRepo.FlushCache(c)
}
