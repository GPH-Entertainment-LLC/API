package service

import (
	"context"
	"xo-packs/core"
	"xo-packs/model"
	"xo-packs/repository"
)

type ReportService interface {
	GetReportOpts(context.Context) ([]*model.ReportOpt, error)
	SubmitReport(context.Context, *model.Report) (*model.ReportExpanded, error)
}

type ReportSvcImpl struct {
	reportRepo repository.ReportRepository
}

func NewReportService(repo repository.ReportRepository) ReportService {
	return &ReportSvcImpl{reportRepo: repo}
}

func (service *ReportSvcImpl) GetReportOpts(c context.Context) ([]*model.ReportOpt, error) {
	return service.reportRepo.GetReportOpts(c)
}

func (service *ReportSvcImpl) SubmitReport(c context.Context, report *model.Report) (*model.ReportExpanded, error) {
	previousReport, err := service.reportRepo.GetSubmittedReport(c, *report.Uid, *report.ReportedUid)
	if err != nil {
		return nil, err
	}
	if previousReport.ID != nil {
		return nil, &core.ErrorResp{Message: "Report already submitted for this user"}
	}

	return service.reportRepo.SubmitReport(c, report)
}
