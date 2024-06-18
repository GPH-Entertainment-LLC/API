package controller

import (
	"fmt"
	"net/http"
	"strings"
	"xo-packs/core"
	"xo-packs/db"
	"xo-packs/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/swaggo/swag/example/celler/httputil"
)

type LoggingController struct {
	loggingService service.LoggingService
	userService    service.UserService
}

func NewLoggingService(loggingService service.LoggingService, userService service.UserService) *LoggingController {
	return &LoggingController{loggingService: loggingService, userService: userService}
}

func (contr LoggingController) Register(router *gin.Engine) {
	router.POST("/log/signIn/:uid", contr.LogSignIn)
	router.POST("/log/ageAgreement/:uid", contr.LogAgeAgreement)
}

// @Summary			Post a user sign in log
// @Description		Log a users sign in to persistent data store
// @Accept 			json
// @Produce			json
// @Param			uid path string true "uid"
// @Tags			Logging
// @Success			200 {object} model.SignInLog
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/log/signIn/:uid [post]
func (contr LoggingController) LogSignIn(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "uid param must be present"})
		return
	}

	user, err := contr.userService.GetUser(c.Request.Context(), uid, true)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	if user == nil {
		httputil.NewError(c, http.StatusInternalServerError, &core.ErrorResp{Message: "user does not exist"})
		return
	}
	if user.Uid == nil {
		httputil.NewError(c, http.StatusInternalServerError, &core.ErrorResp{Message: "user uid does not exist"})
		return
	}
	if user.Email == nil {
		httputil.NewError(c, http.StatusInternalServerError, &core.ErrorResp{Message: fmt.Sprintf("data quality error: user %v does not have email", uid)})
		return
	}
	if user.Username == nil {
		httputil.NewError(c, http.StatusInternalServerError, &core.ErrorResp{Message: fmt.Sprintf("data quality error: user %v does not have username", uid)})
		return
	}

	// log user sign in
	core.AddLog(logrus.Fields{
		"UID":      uid,
		"Username": *user.Username,
		"Email":    *user.Email,
	}, c, db.LOG_SIGN_IN)

	// log in DB
	forwardedHeader := c.Request.Header["Forwarded"]
	forwardedValues := []string{}
	if len(forwardedHeader) > 0 {
		forwardedValues = strings.Split(forwardedHeader[0], ";")
	}

	clientIP := ""
	if len(forwardedValues) > 0 {
		clientIP = forwardedValues[0]
		clientIP = strings.TrimPrefix(clientIP, "for=")
	}
	log, err := contr.loggingService.LogSignIn(c, uid, clientIP)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, log)
	return
}

func (contr LoggingController) LogAgeAgreement(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "uid param must be present"})
		return
	}

	user, err := contr.userService.GetUser(c.Request.Context(), uid, true)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	if user == nil {
		httputil.NewError(c, http.StatusInternalServerError, &core.ErrorResp{Message: "user does not exist"})
		return
	}
	if user.Uid == nil {
		httputil.NewError(c, http.StatusInternalServerError, &core.ErrorResp{Message: "user uid does not exist"})
		return
	}
	if user.Email == nil {
		httputil.NewError(c, http.StatusInternalServerError, &core.ErrorResp{Message: fmt.Sprintf("data quality error: user %v does not have email", uid)})
		return
	}
	if user.Username == nil {
		httputil.NewError(c, http.StatusInternalServerError, &core.ErrorResp{Message: fmt.Sprintf("data quality error: user %v does not have username", uid)})
		return
	}

	// log in DB
	forwardedHeader := c.Request.Header["Forwarded"]
	forwardedValues := []string{}
	if len(forwardedHeader) > 0 {
		forwardedValues = strings.Split(forwardedHeader[0], ";")
	}

	clientIP := ""
	if len(forwardedValues) > 0 {
		clientIP = forwardedValues[0]
		clientIP = strings.TrimPrefix(clientIP, "for=")
	}
	log, err := contr.loggingService.LogAgeAgreement(c, uid, clientIP)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, log)
	return
}
