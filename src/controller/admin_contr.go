package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"xo-packs/core"
	"xo-packs/db"
	"xo-packs/model"
	"xo-packs/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/swaggo/swag/example/celler/httputil"
)

type AdminController struct {
	userService        service.UserService
	applicationService service.ApplicationService
	adminService       service.AdminService
}

func NewAdminController(
	userService service.UserService,
	applicationService service.ApplicationService,
	adminService service.AdminService,
) *AdminController {
	return &AdminController{
		userService:        userService,
		applicationService: applicationService,
		adminService:       adminService,
	}
}

func (contr AdminController) Register(router *gin.Engine) {
	router.POST("/admin/login", contr.AdminLogin)
	router.POST("/admin/approveCreator", contr.ApproveVendor)
	router.POST("/admin/rejectCreator", contr.RejectVendor)
	router.POST("/admin/addFaq", contr.AddFaq)
	router.PATCH("/admin/editFaq", contr.EditFaq)
	router.DELETE("/admin/removeFaq", contr.RemoveFaq)
	router.DELETE("/admin/removeCreator", contr.RemoveVendor)
	router.DELETE("/admin/cache/flush", contr.FlushCache)
}

// @Summary			Login as an admin
// @Description		Authenticate and login to the admin dashboard
// @Param			authorizedUid query string true "authorized uid"
// @Accept			json
// @Produce			json
// @Tags			Admin
// @Success			200 {} string
// @Failure 		401 {object} httputil.HTTPError
// @Router			/admin/login [POST]
func (contr AdminController) AdminLogin(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{Message: "valid uid must be present"})
		return
	}

	fmt.Println("INSIDE ADMIN LOGIN")

	// log vendor approval
	err := core.AddLog(logrus.Fields{
		"AdminUid": authorizedUid,
	}, c, db.LOG_ADMIN_LOG)

	if err != nil {
		fmt.Println("Error could not log: ", err)
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, "success")
	return
}

// @Summary			Approve a new vendors application
// @Description		Approve a vendors application and verify their account
// @Param			authorizedUid query string true "authorized uid"
// @Param	 		vendorUid query string true "creator uid"
// @Accept			json
// @Produce			json
// @Tags			Admin
// @Success			200 {object} model.VendorApplication
// @Failure 		401 {object} httputil.HTTPError
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/admin/approveVendor [POST]
func (contr AdminController) ApproveVendor(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{Message: "valid uid must be present"})
		return
	}

	vendorUid := c.Query("creatorUid")
	if vendorUid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{
			Message: "a vendor uid param must be present",
		})
		return
	}

	user, err := contr.userService.GetUser(c.Request.Context(), vendorUid, true)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	if user == nil ||
		user.Uid == nil ||
		user.Email == nil ||
		user.Username == nil {
		httputil.NewError(c, http.StatusInternalServerError, &core.ErrorResp{
			Message: "Data quality error: user does not have all required fields",
		})
		return
	}

	vendorApplication, err := contr.adminService.ApproveVendor(c.Request.Context(), *user.Uid, *user.Username, contr.applicationService, contr.userService)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	if vendorApplication.FirstName == nil ||
		vendorApplication.LastName == nil ||
		vendorApplication.Uid == nil ||
		vendorApplication.Email == nil ||
		vendorApplication.Username == nil {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{
			Message: "Creator Application data quality error: not all required fields exist in application",
		})
		return
	}

	creatorAccountPatchMap := map[string]interface{}{
		"firstName": *vendorApplication.FirstName,
		"lastName":  *vendorApplication.LastName,
	}

	patchedUser, err := contr.userService.PatchUser(c, *vendorApplication.Uid, *vendorApplication.Username, creatorAccountPatchMap)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
	}

	fmt.Println("=== Patched Approved Creator: ", patchedUser)

	// log creator approval
	core.AddLog(logrus.Fields{
		"UID":       *vendorApplication.Uid,
		"AdminUid":  authorizedUid,
		"Username":  *vendorApplication.Username,
		"Email":     *vendorApplication.Email,
		"Status":    "approved",
		"FirstName": *vendorApplication.FirstName,
		"LastName":  *vendorApplication.LastName,
	}, c, db.LOG_APPLICATION_STATUS)

	c.JSON(http.StatusOK, vendorApplication)
	return
}

// @Summary			Reject a vendor application
// @Description		Reject a vendorss application
// @Param			authorizedUid query string true "authorized uid"
// @Param			vendorUid query string true "creator uid"
// @Accept			json
// @Produce			json
// @Tags			Admin
// @Success			200 {object} model.VendorApplication
// @Failure 		401 {object} httputil.HTTPError
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/admin/rejectVendor [POST]
func (contr AdminController) RejectVendor(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{Message: "valid uid must be present"})
		return
	}

	vendorUid := c.Query("creatorUid")
	if vendorUid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{
			Message: "a creator uid param must be present",
		})
		return
	}

	user, err := contr.userService.GetUser(c.Request.Context(), vendorUid, true)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	if user == nil ||
		user.Uid == nil ||
		user.Email == nil ||
		user.Username == nil {
		httputil.NewError(c, http.StatusInternalServerError, &core.ErrorResp{
			Message: "Data quality error: user does not have all required fields",
		})
		return
	}

	vendorApplication, err := contr.adminService.RejectVendor(c.Request.Context(), *user.Uid, *user.Username, contr.applicationService)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	if vendorApplication.FirstName == nil ||
		vendorApplication.LastName == nil ||
		vendorApplication.Uid == nil ||
		vendorApplication.Email == nil ||
		vendorApplication.Username == nil {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{
			Message: "Creator Application data quality error: not all required fields exist in application",
		})
		return
	}

	// log creator rejection
	core.AddLog(logrus.Fields{
		"UID":       *vendorApplication.Uid,
		"AdminUid":  authorizedUid,
		"Username":  *vendorApplication.Username,
		"Email":     *vendorApplication.Email,
		"Status":    "rejected",
		"FirstName": *vendorApplication.FirstName,
		"LastName":  *vendorApplication.LastName,
	}, c, db.LOG_APPLICATION_STATUS)

	c.JSON(http.StatusOK, vendorApplication)
	return
}

// @Summary			Remove a vendor
// @Description		Revoke a vendors verification
// @Param			authorizedUid query string true "authorized uid"
// @Param			vendorUid query string true "creator uid"
// @Accept			json
// @Produce			json
// @Tags			Admin
// @Success			200 {} string
// @Failure 		401 {object} httputil.HTTPError
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/admin/removeCreator [DELETE]
func (contr AdminController) RemoveVendor(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "valid uid must be present"})
		return
	}

	vendorUid := c.Query("creatorUid")
	if vendorUid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{
			Message: "a creator uid param must be present",
		})
		return
	}

	user, err := contr.userService.GetUser(c.Request.Context(), vendorUid, true)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	if user == nil ||
		user.Uid == nil ||
		user.Email == nil ||
		user.Username == nil {
		httputil.NewError(c, http.StatusInternalServerError, &core.ErrorResp{
			Message: "Data quality error: user does not have all required fields",
		})
		return
	}

	err = contr.adminService.RemoveVendor(c.Request.Context(), vendorUid, *user.Username, contr.userService)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	// log creator removal
	core.AddLog(logrus.Fields{
		"UID":      vendorUid,
		"AdminUid": authorizedUid,
		"Username": *user.Username,
		"Email":    *user.Email,
	}, c, db.LOG_VENDOR_REMOVAL)

	c.JSON(http.StatusOK, "success")
	return
}

// @Summary			Add an faq
// @Description		Adds an faq to the list of active faq
// @Param			authorizedUid query string true "authorized uid"
// @Param			faq body model.Faq true "faq"
// @Accept			json
// @Produce			json
// @Tags			Admin
// @Success			200 {object} model.Faq
// @Failure 		401 {object} httputil.HTTPError
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/admin/addFaq [POST]
func (contr AdminController) AddFaq(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "valid uid must be present"})
		return
	}

	faq := new(model.Faq)
	if err := c.BindJSON(faq); err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	faq, err := contr.adminService.AddFaq(c, faq)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, faq)
	return
}

// @Summary			Remove an faq
// @Description		Remoe an faq from the active list
// @Param			authorizedUid query string true "authorized uid"
// @Param			id query int true "faq id"
// @Accept			json
// @Produce			json
// @Tags			Admin
// @Success			200 {} string
// @Failure 		401 {object} httputil.HTTPError
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/admin/removeFaq [DELETE]
func (contr AdminController) RemoveFaq(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "valid uid must be present"})
		return
	}

	rawId := c.Query("id")
	id, err := strconv.Atoi(rawId)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	if err = contr.adminService.RemoveFaq(c, uint64(id)); err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, "sucess")
	return
}

// @Summary			Edit an existing faq
// @Description		Changes the structure of an existing faq
// @Param			authorizedUid query string true "authorized uid"
// @Param			id query int true "faq id"
// @Param			faqPatch body model.Faq true "faq patch map"
// @Accept			json
// @Produce			json
// @Tags			Admin
// @Success			200 {} string
// @Failure 		401 {object} httputil.HTTPError
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/admin/editFaq [PATCH]
func (contr AdminController) EditFaq(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "valid uid must be present"})
		return
	}

	rawId := c.Query("id")
	id, err := strconv.Atoi(rawId)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	patchMap := map[string]interface{}{}
	if err := c.BindJSON(&patchMap); err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	err = contr.adminService.EditFaq(c, patchMap, uint64(id))
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, "success")
	return
}

func (contr AdminController) FlushCache(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "valid uid must be present"})
		return
	}

	err := contr.adminService.FlushCache(c)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, "success")
	return
}
