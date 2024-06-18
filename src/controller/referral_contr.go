package controller

import (
	"net/http"
	"xo-packs/core"
	"xo-packs/service"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/example/celler/httputil"
)

type ReferralController struct {
	referralService service.ReferralService
	vendorService   service.VendorService
}

func NewReferralController(referralService service.ReferralService, vendorService service.VendorService) *ReferralController {
	return &ReferralController{referralService: referralService, vendorService: vendorService}
}

func (contr ReferralController) Register(router *gin.Engine) {
	router.POST("/referral/generate", contr.GenerateCode)
	router.POST("/referral/create", contr.CreateCode)
	router.DELETE("/referral/remove", contr.RemoveCode)
	router.GET("/referral/getActiveCodes", contr.GetActiveCodes)
}

// @Summary 		Generate a new referral code
// @Description 	A creator can generate a new referral code to use for referring new creators
// @Param			authorizedUid query string true "authorized uid"
// @Tags 			Referral
// @Accept 			json
// @Produce 		json
// @Success 		201 {object} model.ReferralCode
// @Failure 		500 {object} httputil.HTTPError
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		401 {object} httputil.HTTPError
// @Router 			/referral/generate [POST]
func (contr ReferralController) GenerateCode(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "a valid uid pararm must be present",
		})
		return
	}

	referralCode, err := contr.referralService.GenerateCode(c.Request.Context(), authorizedUid)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, referralCode)
	return
}

// @Summary 		Create a custom code referral code
// @Description 	A creator can upload a new custom referral code to use for referring new creators
// @Param			authorizedUid query string true "authorized uid"
// @Param			code query string true "referral code"
// @Tags 			Referral
// @Accept 			json
// @Produce 		json
// @Success 		201 {object} model.ReferralCode
// @Failure 		500 {object} httputil.HTTPError
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		401 {object} httputil.HTTPError
// @Router 			/referral/create [POST]
func (contr ReferralController) CreateCode(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "a valid uid pararm must be present",
		})
		return
	}

	// code must be alphanumeric at most 40 chars long
	code := c.Query("code")
	if code == "" {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "a valid code must be present",
		})
		return
	}

	referralCode, err := contr.referralService.CreateCode(c.Request.Context(), authorizedUid, code)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, referralCode)
	return
}

// @Summary 		Remove a referral code
// @Description 	A creator can deactivate a referral code so it can no longer be used
// @Param			authorizedUid query string true "authorized uid"
// @Param			code query string true "referral code"
// @Tags 			Referral
// @Accept 			json
// @Produce 		json
// @Success 		200 {} string
// @Failure 		500 {object} httputil.HTTPError
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		401 {object} httputil.HTTPError
// @Router 			/referral/remove [DELETE]
func (contr ReferralController) RemoveCode(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "a valid uid pararm must be present",
		})
		return
	}

	// code must be alphanumeric at most 40 chars long
	code := c.Query("code")
	if code == "" {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "a valid code must be present",
		})
		return
	}

	err := contr.referralService.RemoveCode(c.Request.Context(), authorizedUid, code)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, "success")
	return
}

// @Summary 		Get all active codes that a creator has
// @Description 	A creator can get a list of all active referral codes associated with them
// @Param			authorizedUid query string true "authorized uid"
// @Tags 			Referral
// @Accept 			json
// @Produce 		json
// @Success 		200 {object} []model.ReferralCode
// @Failure 		500 {object} httputil.HTTPError
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		401 {object} httputil.HTTPError
// @Router 			/referral/getActiveCodes [GET]
func (contr ReferralController) GetActiveCodes(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "a valid uid pararm must be present",
		})
		return
	}

	referralCodes, err := contr.referralService.GetActiveCodes(c.Request.Context(), authorizedUid)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, referralCodes)
	return
}
