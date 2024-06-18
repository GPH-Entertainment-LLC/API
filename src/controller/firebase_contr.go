package controller

import (
	"fmt"
	"io"
	"net/http"
	"xo-packs/core"
	"xo-packs/db"
	"xo-packs/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/swaggo/swag/example/celler/httputil"
)

type FirebaseController struct {
	firebaseService service.FirebaseService
	userService     service.UserService
}

func NewFirebaseController(firebaseService service.FirebaseService, userService service.UserService) *FirebaseController {
	return &FirebaseController{firebaseService: firebaseService, userService: userService}
}

func (contr FirebaseController) Register(router *gin.Engine) {
	router.GET("/firebase/getUserByEmail", contr.GetFirebaseUserByEmail)
	router.DELETE("/firebase/deleteUser", contr.DeleteFirebaseUser)
}

func (contr FirebaseController) GetFirebaseUserByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{
			Message: "email param must be present",
		})
		return
	}

	resp, err := contr.firebaseService.GetUserByEmail(c, email)
	if err != nil {
		c.JSON(resp.StatusCode, err.Error())
		return
	} else {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(resp.StatusCode, err.Error())
			return
		}
		c.IndentedJSON(resp.StatusCode, string(body))
		return
	}
}

func (contr FirebaseController) DeleteFirebaseUser(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "valid uid must be present"})
		return
	}

	username := c.Query("username")
	if username == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "username param must be present"})
		return
	}

	user, err := contr.userService.GetUser(c.Request.Context(), authorizedUid, true)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	if user == nil {
		httputil.NewError(c, http.StatusInternalServerError, &core.ErrorResp{Message: "user does not exist"})
		return
	}
	if user.Uid == nil {
		httputil.NewError(c, http.StatusInternalServerError, &core.ErrorResp{Message: "user does not exist"})
		return
	}
	if user.Email == nil {
		httputil.NewError(c, http.StatusInternalServerError, &core.ErrorResp{Message: fmt.Sprintf("data quality error: user %v does not have email", authorizedUid)})
		return
	}
	if user.Username == nil {
		httputil.NewError(c, http.StatusInternalServerError, &core.ErrorResp{Message: fmt.Sprintf("data quality error: user %v does not have username", authorizedUid)})
		return
	}

	resp, err := contr.firebaseService.DeleteUser(c, authorizedUid, username, contr.userService)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	} else {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			httputil.NewError(c, http.StatusInternalServerError, err)
			return
		}

		// log user deletion
		core.AddLog(logrus.Fields{
			"UID":      *user.Uid,
			"Username": *user.Username,
			"Email":    *user.Email,
		}, c, db.LOG_USER_DELETE)

		c.JSON(resp.StatusCode, body)
		return
	}
}
