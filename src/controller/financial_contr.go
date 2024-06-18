package controller

import (
	"net/http"
	"xo-packs/core"
	"xo-packs/service"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/example/celler/httputil"
)

type FinancialController struct {
	financialService service.FinancialService
}

func NewFinancialController(financialService service.FinancialService) *FinancialController {
	return &FinancialController{financialService: financialService}
}

func (contr FinancialController) Register(router *gin.Engine) {
	router.GET("/financial/creatorEarnings/:creatorUid", contr.GetCreatorEarnings)
	router.GET("/financial/referralEarnings/:creatorUid", contr.GetReferralEarnings)
	router.GET("/financial/allEarnings/:creatorUid", contr.GetAllEarnings)
}

// @Summary			Get list of creator earnings
// @Description		Get list of creator earnings for a creator
// @Accept			json
// @Produce			json
// @Tags			Financial
// @Success			200 {object} []model.CreatorEarningsPeriod
// @Failure 		500 {object} httputil.HTTPError
// @Router			/financial/creatorEarnings/:creatorUid [get]
func (contr FinancialController) GetCreatorEarnings(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{Message: "valid uid must be present"})
		return
	}

	creatorUid := c.Param("creatorUid")
	if creatorUid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{
			Message: "a creator uid param must be present",
		})
		return
	}

	if creatorUid != authorizedUid {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "user is unauthorized to perform this action",
		})
		return
	}

	creatorEarnings, err := contr.financialService.GetCreatorEarnings(c, creatorUid)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, creatorEarnings)
	return
}

// @Summary			Get list of referral earnings
// @Description		Get list of referral earnings for creator
// @Accept			json
// @Produce			json
// @Tags			Financial
// @Success			200 {object} []model.ReferralEarningsPeriod
// @Failure 		500 {object} httputil.HTTPError
// @Router			/financial/referralEarnings/:creatorUid [get]
func (contr FinancialController) GetReferralEarnings(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{Message: "valid uid must be present"})
		return
	}

	creatorUid := c.Param("creatorUid")
	if creatorUid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{
			Message: "a creator uid param must be present",
		})
		return
	}

	if creatorUid != authorizedUid {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "user is unauthorized to perform this action",
		})
		return
	}

	referralEarnings, err := contr.financialService.GetReferralEarnings(c, creatorUid)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, referralEarnings)
	return
}

// @Summary			Get list of all earnings
// @Description		Get list of all earnings for creator
// @Accept			json
// @Produce			json
// @Tags			Financial
// @Success			200 {object} []model.AllEarningsPeriod
// @Failure 		500 {object} httputil.HTTPError
// @Router			/financial/allEarnings/:creatorUid [get]
func (contr FinancialController) GetAllEarnings(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{Message: "valid uid must be present"})
		return
	}

	creatorUid := c.Param("creatorUid")
	if creatorUid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{
			Message: "a creator uid param must be present",
		})
		return
	}

	if creatorUid != authorizedUid {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "user is unauthorized to perform this action",
		})
		return
	}

	allEarnings, err := contr.financialService.GetAllEarnings(c, creatorUid)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, allEarnings)
	return
}
