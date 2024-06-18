package controller

import (
	"net/http"
	"xo-packs/core"
	"xo-packs/db"
	"xo-packs/model"
	"xo-packs/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/swaggo/swag/example/celler/httputil"
)

type ReportController struct {
	reportService service.ReportService
}

func NewReportController(reportService service.ReportService) *ReportController {
	return &ReportController{reportService: reportService}
}

func (contr ReportController) Register(router *gin.Engine) {
	router.GET("/report/opts", contr.GetReportOpts)
	router.POST("/report/submit", contr.SubmitReport)
}

// @Summary 		Get active report options
// @Description 	A user can get a list of the available report options
// @Param			authorizedUid query string true "authorized uid"
// @Tags 			Report
// @Accept 			json
// @Produce 		json
// @Success 		200 {object} []model.ReportOpt
// @Failure 		500 {object} httputil.HTTPError
// @Failure 		401 {object} httputil.HTTPError
// @Router 			/report/opts [GET]
func (contr ReportController) GetReportOpts(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "valid uid param must be present",
		})
		return
	}

	reportOpts, err := contr.reportService.GetReportOpts(c.Request.Context())
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, reportOpts)
	return
}

// @Summary 		Submit a new report
// @Description 	A user can submit a report against another user
// @Param			authorizedUid query string true "authorized uid"
// @Param			report body model.Report true "report object"
// @Tags 			Report
// @Accept 			json
// @Produce 		json
// @Success 		201 {object} model.ReportExpanded
// @Failure 		500 {object} httputil.HTTPError
// @Failure 		401 {object} httputil.HTTPError
// @Router 			/report/submit [POST]
func (contr ReportController) SubmitReport(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "valid uid param must be present",
		})
		return
	}

	report := model.Report{}
	err := c.BindJSON(&report)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{
			Message: "Unable to bind report",
		})
		return
	}
	if report.ReportedUid == nil || report.Uid == nil || report.OptId == nil {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{
			Message: "reportedUid, uid, and optId must be present in payload",
		})
		return
	}

	submittedReport, err := contr.reportService.SubmitReport(c.Request.Context(), &report)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	// log report
	notes := ""
	if submittedReport.Notes != nil {
		notes = *submittedReport.Notes
	}
	core.AddLog(logrus.Fields{
		"UID":         *submittedReport.Uid,
		"ReportedUid": *submittedReport.ReportedUid,
		"ReportedAt":  *submittedReport.ReportedAt,
		"ReportOpt":   *submittedReport.Opt,
		"Notes":       notes,
	}, c, db.LOG_REPORT)

	c.JSON(http.StatusCreated, submittedReport)
	return
}
