package service

import (
	"context"
	"xo-packs/core"
	"xo-packs/model"
	"xo-packs/repository"
)

type AnalyticsService interface {
	TotalVendorPacksSold(context.Context, string) (*uint64, error)
	TotalRevenueGenerated(context.Context, string) (*float64, error)
	FavoriteAmount(context.Context, string) (*uint64, error)
	AvgPackQtyPurchased(context.Context, string) (*float64, error)
	MinPackQtyPurchased(context.Context, string) (*uint64, error)
	MaxPackQtyPurchased(context.Context, string) (*uint64, error)
	TopCustomers(context.Context, string) ([]*model.CustomerAnalytic, error)
	PackSales(context.Context, string, string, int64, int64, string) (*model.PackAnalyticsResp, error)
	PackQtySold(context.Context, string) ([]*model.PackQtySoldAnalytic, error)
}

type AnalyticsSvcImpl struct {
	analyticsRepo repository.AnalyticsRepository
}

func NewAnalyticsService(repo repository.AnalyticsRepository) AnalyticsService {
	return &AnalyticsSvcImpl{analyticsRepo: repo}
}

func (analyticsService AnalyticsSvcImpl) TotalVendorPacksSold(c context.Context, vendorId string) (*uint64, error) {
	return analyticsService.analyticsRepo.TotalVendorPacksSold(c, vendorId)
}

func (analyticsService AnalyticsSvcImpl) TotalRevenueGenerated(c context.Context, vendorId string) (*float64, error) {
	return analyticsService.analyticsRepo.TotalRevenueGenerated(c, vendorId)
}

func (analyticsService AnalyticsSvcImpl) FavoriteAmount(c context.Context, vendorId string) (*uint64, error) {
	return analyticsService.analyticsRepo.FavoriteAmount(c, vendorId)
}

func (analyticsService AnalyticsSvcImpl) AvgPackQtyPurchased(c context.Context, vendorId string) (*float64, error) {
	return analyticsService.analyticsRepo.AvgPackQtyPurchased(c, vendorId)
}

func (analyticsService AnalyticsSvcImpl) MinPackQtyPurchased(c context.Context, vendorId string) (*uint64, error) {
	return analyticsService.analyticsRepo.MinPackQtyPurchased(c, vendorId)
}

func (analyticsService AnalyticsSvcImpl) MaxPackQtyPurchased(c context.Context, vendorId string) (*uint64, error) {
	return analyticsService.analyticsRepo.MaxPackQtyPurchased(c, vendorId)
}

func (analyticsService AnalyticsSvcImpl) TopCustomers(c context.Context, vendorId string) ([]*model.CustomerAnalytic, error) {
	return analyticsService.analyticsRepo.TopCustomers(c, vendorId)
}

func (analyticsService AnalyticsSvcImpl) PackSales(c context.Context, vendorId string, granularity string, month int64, year int64, urlPath string) (*model.PackAnalyticsResp, error) {
	var granularityStrs []string
	switch granularity {
	case "day":
		granularityStrs = core.GenerateDaysOfMonth(year, month)
	case "month":
		granularityStrs = core.GenerateMonthStrings()
	case "year":
		granularityStrs = core.GenerateLast5Years()
	}

	packSales, err := analyticsService.analyticsRepo.PackSales(c, vendorId, granularity, month, year, urlPath)
	if err != nil {
		return nil, err
	}
	return &model.PackAnalyticsResp{
		TimeAxis: granularityStrs,
		DataSet:  packSales,
	}, nil
}

func (analyticsService AnalyticsSvcImpl) PackQtySold(c context.Context, vendorId string) ([]*model.PackQtySoldAnalytic, error) {
	return analyticsService.analyticsRepo.PackQtySold(c, vendorId)
}
