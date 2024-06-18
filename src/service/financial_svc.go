package service

import (
	"context"
	"xo-packs/model"
	"xo-packs/repository"
)

type FinancialService interface {
	GetCreatorEarnings(context.Context, string) ([]*model.CreatorEarningsPeriod, error)
	GetReferralEarnings(context.Context, string) ([]*model.ReferralEarningsPeriod, error)
	GetAllEarnings(context.Context, string) ([]*model.AllEarningsPeriod, error)
}

type FinancialSvcImpl struct {
	financialRepo repository.FinancialRepository
}

func NewFinancialService(repo repository.FinancialRepository) FinancialService {
	return &FinancialSvcImpl{financialRepo: repo}
}

func (financialService FinancialSvcImpl) GetCreatorEarnings(c context.Context, creatorUid string) ([]*model.CreatorEarningsPeriod, error) {
	return financialService.financialRepo.GetCreatorEarnings(c, creatorUid)
}

func (financialService FinancialSvcImpl) GetReferralEarnings(c context.Context, creatorUid string) ([]*model.ReferralEarningsPeriod, error) {
	return financialService.financialRepo.GetReferralEarnings(c, creatorUid)
}

func (financialService FinancialSvcImpl) GetAllEarnings(c context.Context, creatorUid string) ([]*model.AllEarningsPeriod, error) {
	return financialService.financialRepo.GetAllEarnings(c, creatorUid)
}
