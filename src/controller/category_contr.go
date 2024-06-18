package controller

import (
	"net/http"
	"xo-packs/core"
	"xo-packs/service"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/example/celler/httputil"
)

type CategoryController struct {
	categoryService service.CategoryService
}

func NewCategoryController(categoryService service.CategoryService) *CategoryController {
	return &CategoryController{categoryService: categoryService}
}

func (contr CategoryController) Register(router *gin.Engine) {
	router.GET("/categories/all", contr.GetAllCategories)
	router.GET("/categories/:category", contr.GetCategories)
}

// @Summary			Get list of all categories
// @Description		Get list of all categories
// @Accept			json
// @Produce			json
// @Tags			Category
// @Success			200 {object} []model.Category
// @Failure 		500 {object} httputil.HTTPError
// @Router			/categories/all [get]
func (contr CategoryController) GetAllCategories(c *gin.Context) {
	categories, err := contr.categoryService.GetAllCategories(c.Request.Context())
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, categories)
	return
}

// @Summary			Get list of specific categories
// @Description		Get list of specific categories
// @Param			category path string true "category param must be item,pack,or vendor"
// @Accept			json
// @Produce			json
// @Tags			Category
// @Success			200 {object} []model.Category
// @Failure 		500 {object} httputil.HTTPError
// @Router			/categories/:category [get]
func (contr CategoryController) GetCategories(c *gin.Context) {
	category := c.Param("category")
	if category == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "category path param must be present"})
		return
	}

	categories, err := contr.categoryService.GetCategories(c.Request.Context(), category)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, categories)
	return
}
