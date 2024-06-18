package controller

import (
	"net/http"
	"xo-packs/core"
	"xo-packs/service"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/example/celler/httputil"
)

type ShippingController struct {
	shippingService service.ShippingService
	userService     service.UserService
}

func NewShippingController(
	shippingService service.ShippingService,
	userService service.UserService,
) *ShippingController {
	return &ShippingController{
		shippingService: shippingService,
		userService:     userService,
	}
}

func (contr ShippingController) Register(router *gin.Engine) {
	router.GET("/shippingInfo/:uid", contr.GetShippingInfo)
	router.POST("/shippingInfo", contr.UploadShippingInfo)
	router.PATCH("/shippingInfo/:uid", contr.UpdateShippingInfo)
}

func (contr ShippingController) GetShippingInfo(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{
			Message: "valid uid param must be present",
		})
		return
	}

}

func (contr ShippingController) UploadShippingInfo(c *gin.Context) {

}

func (contr ShippingController) UpdateShippingInfo(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{
			Message: "valid uid param must be present",
		})
		return
	}

}
