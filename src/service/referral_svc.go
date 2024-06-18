package service

import (
	"context"
	"xo-packs/core"
	"xo-packs/model"
	"xo-packs/repository"
)

type ReferralService interface {
	GenerateCode(context.Context, string) (*model.ReferralCode, error)
	CreateCode(context.Context, string, string) (*model.ReferralCode, error)
	RemoveCode(context.Context, string, string) error
	ValidateCode(context.Context, string, string) (*model.Referral, error)
	GetActiveCodes(context.Context, string) ([]*model.ReferralCode, error)
}

type ReferralSvcImpl struct {
	referralRepo repository.ReferralRepository
}

func NewReferralService(repo repository.ReferralRepository) ReferralService {
	return &ReferralSvcImpl{referralRepo: repo}
}

func (service *ReferralSvcImpl) GenerateCode(c context.Context, uid string) (*model.ReferralCode, error) {
	code := ""
	for attempt := 0; attempt < 5; attempt++ {
		code, err := core.GenerateReferralCode()
		if err != nil {
			return nil, err
		}
		referralCode, err := service.referralRepo.GetReferralCode(c, code)
		if err != nil {
			return nil, err
		}
		if referralCode.Code == nil {
			break
		}
	}
	if code == "" {
		return nil, &core.ErrorResp{Message: "An unexpected error occurred generating a new code. Please try again later"}
	}
	if err := service.referralRepo.ClearReferralCache(c, uid); err != nil {
		return nil, err
	}
	return service.referralRepo.CreateCode(c, uid, code)
}

func (service *ReferralSvcImpl) CreateCode(c context.Context, uid string, code string) (*model.ReferralCode, error) {
	activeCodes, err := service.referralRepo.GetActiveCodes(c, uid)
	if err != nil {
		return nil, err
	}

	if len(activeCodes) >= 5 {
		return nil, &core.ErrorResp{Message: "cannot have more than 5 active codes at a time"}
	}

	referralCode, err := service.referralRepo.GetReferralCode(c, code)
	if err != nil {
		return nil, err
	}

	if referralCode.Code != nil {
		return nil, &core.ErrorResp{Message: "code already in use, please use a different one"}
	}
	if err = service.referralRepo.ClearReferralCache(c, uid); err != nil {
		return nil, err
	}
	if err = service.referralRepo.ClearReferralCodeCache(c, code); err != nil {
		return nil, err
	}
	return service.referralRepo.CreateCode(c, uid, code)
}

func (service *ReferralSvcImpl) RemoveCode(c context.Context, uid string, code string) error {
	if err := service.referralRepo.ClearReferralCache(c, uid); err != nil {
		return err
	}
	if err := service.referralRepo.ClearReferralCodeCache(c, uid); err != nil {
		return err
	}
	return service.referralRepo.RemoveCode(c, uid, code)
}

func (service *ReferralSvcImpl) ValidateCode(c context.Context, refereeUid string, code string) (*model.Referral, error) {
	referral, err := service.referralRepo.GetReferralCode(c, code)
	if err != nil {
		return nil, err
	}

	if referral.Code == nil || referral.Uid == nil {
		return nil, &core.ErrorResp{Message: "referral code does not exist"}
	}
	return service.referralRepo.ValidateCode(c, *referral.Uid, refereeUid, code)
}

func (service *ReferralSvcImpl) GetActiveCodes(c context.Context, uid string) ([]*model.ReferralCode, error) {
	return service.referralRepo.GetActiveCodes(c, uid)
}
