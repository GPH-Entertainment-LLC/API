package controller

import (
	"encoding/json"
	"net/http"
	"xo-packs/core"
	"xo-packs/db"
	"xo-packs/model"
	"xo-packs/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/swaggo/swag/example/celler/httputil"
)

type ApplicationController struct {
	applicationService service.ApplicationService
	referralService    service.ReferralService
}

func NewApplicationController(applicationService service.ApplicationService, referralService service.ReferralService) *ApplicationController {
	return &ApplicationController{applicationService: applicationService, referralService: referralService}
}

func (contr ApplicationController) Register(router *gin.Engine) {
	router.POST("/creator/application", contr.VendorApplicationSubmit)
}

// @Summary 		Post a new vendor application
// @Description 	A user submits a new vendor application
// @Param			authorizedUid query string true "authorized uid"
// @Tags 			Application
// @Accept 			json
// @Produce 		json
// @Success 		200 {object} model.VendorApplication
// @Failure 		500 {object} httputil.HTTPError
// @Failure 		400 {object} httputil.HTTPError
// @Router 			/application [POST]
func (contr ApplicationController) VendorApplicationSubmit(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "valid uid param must be present",
		})
		return
	}

	// get front ID image file
	_, frontIdHeader, err := c.Request.FormFile("frontIdFile")
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	// get back ID image file
	_, backIdHeader, err := c.Request.FormFile("backIdFile")
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	// get profile ID image file
	_, profileIdHeader, err := c.Request.FormFile("profileIdFile")
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	rawVendorApplication := c.Request.FormValue("creatorApplication")
	vendorApplication := model.VendorApplication{}
	if err = json.Unmarshal([]byte(rawVendorApplication), &vendorApplication); err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	if vendorApplication.Uid == nil ||
		vendorApplication.Email == nil ||
		vendorApplication.Username == nil ||
		vendorApplication.FirstName == nil ||
		vendorApplication.LastName == nil {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{
			Message: "creator application must have uid",
		})
		return
	}

	if vendorApplication.ReferralCode != nil {
		referral, err := contr.referralService.ValidateCode(c.Request.Context(), *vendorApplication.Uid, *vendorApplication.ReferralCode)
		if err != nil {
			httputil.NewError(c, http.StatusInternalServerError, err)
			return
		}
		if referral.ID != nil {
			// log new referral
			core.AddLog(logrus.Fields{
				"UID":         *vendorApplication.Uid,
				"Username":    *vendorApplication.Username,
				"Email":       *vendorApplication.Email,
				"FirstName":   *vendorApplication.FirstName,
				"LastName":    *vendorApplication.LastName,
				"RefereeUid":  *referral.RefereeUid,
				"ReferrerUid": *referral.ReferrerUid,
				"Code":        *referral.Code,
				"ValidatedAt": *referral.ValidatedAt,
			}, c, db.LOG_REFERRAL)
		}
	}

	finalVendorApplication, err := contr.applicationService.ApplicationSubmit(
		c.Request.Context(),
		&vendorApplication,
		frontIdHeader,
		backIdHeader,
		profileIdHeader,
	)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	// log creator application
	notes := ""
	if vendorApplication.Notes != nil {
		notes = *vendorApplication.Notes
	}
	core.AddLog(logrus.Fields{
		"UID":       *vendorApplication.Uid,
		"Username":  *vendorApplication.Username,
		"Email":     *vendorApplication.Email,
		"FirstName": *vendorApplication.FirstName,
		"LastName":  *vendorApplication.LastName,
		"Notes":     notes,
	}, c, db.LOG_APPLICATION)

	c.JSON(http.StatusCreated, finalVendorApplication)
	return
}
