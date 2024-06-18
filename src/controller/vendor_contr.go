package controller

import (
	"net/http"
	"strconv"
	"strings"
	"xo-packs/core"
	"xo-packs/model"
	"xo-packs/service"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/example/celler/httputil"
)

type VendorController struct {
	vendorService   service.VendorService
	categoryService service.CategoryService
	packService     service.PackService
	itemService     service.ItemService
}

func NewVendorController(
	vendorService service.VendorService,
	categoryService service.CategoryService,
	packService service.PackService,
	itemService service.ItemService,
) *VendorController {
	return &VendorController{
		vendorService:   vendorService,
		categoryService: categoryService,
		packService:     packService,
		itemService:     itemService,
	}
}

func (contr *VendorController) Register(router *gin.Engine) {
	router.GET("/vendors/sortlist", contr.GetVendorSortList)
	router.GET("/vendor/:uid", contr.GetVendor)
	router.GET("/vendors", contr.GetVendorsPage)
	router.GET("/vendors/packs", contr.GetVendorShopPackPage)
	router.GET("/vendor/packs/:uid", contr.GetVendorPackListPage)
	router.GET("/vendor/packs/profile/:uid", contr.GetVendorProfilePackPage)
	router.GET("/vendor/items/:uid", contr.GetVendorItemListPage)
	router.GET("/vendor/categories/:vendorId", contr.GetVendorCategories)
	router.POST("/vendor/categories", contr.AddVendorCategories)
	router.DELETE("/vendor/item/clear", contr.ClearVendorItemCache)
	router.DELETE("/vendor/pack/clear", contr.ClearVendorPackCache)
}

// @Summary			Get a vendor
// @Description		Get a vendor given the uid
// @Accept 			json
// @Produce			json
// @Param			uid path string true "uid"
// @Tags			Vendor
// @Success			200 {object} model.Vendor
// @Failure 		500 {object} httputil.HTTPError
// @Router			/vendor/:uid [get]
func (contr *VendorController) GetVendor(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{
			Message: "uid param must be present",
		})
		return
	}

	vendor, err := contr.vendorService.GetActiveVendor(c.Request.Context(), uid)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, vendor)
	return
}

func (contr *VendorController) GetVendorSortList(c *gin.Context) {
	sortList, err := contr.vendorService.GetVendorSortList(c.Request.Context(), c.Request.URL.String())
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, sortList)
	return
}

// @Summary			Get vendor page
// @Description		Get a list of vendors for a page
// @Accept 			json
// @Produce			json
// @Param			pageNum query int true "page number"
// @Param			sortBy query string false "sort by"
// @Param			sortDir query string false "sort direction"
// @Param			filterOn query string false "category filter"
// @Param			search query string false "search string"
// @Tags			Vendor
// @Success			200 {object} model.VendorPage
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/vendors [get]
func (contr *VendorController) GetVendorsPage(c *gin.Context) {
	rawPageNum := c.Query("pageNum")
	pageNum, err := strconv.ParseUint(rawPageNum, 10, 64)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}
	if pageNum <= 0 {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "page number must be positive"})
		return
	}
	searchStr := strings.ToLower(c.Query("search"))
	sortBy := c.Query("sortBy")
	sortDir := c.Query("sortDir")
	filterOn := c.Query("filterOn")

	vendorPage, err := contr.vendorService.GetVendorsPage(
		c.Request.Context(),
		pageNum,
		sortBy,
		sortDir,
		filterOn,
		searchStr,
		c.Request.URL.String(),
		contr.categoryService,
	)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, vendorPage)
	return
}

// @Summary			Get vendor shop pack
// @Description		Get page of vendor shop packs
// @Accept 			json
// @Produce			json
// @Param			pageNum query int true "page number"
// @Param			sortBy query string false "sort by"
// @Param			sortDir query string false "sort direction"
// @Param			filterOn query string false "category filter"
// @Param			search query string false "search string"
// @Tags			Vendor
// @Success			200 {object} model.VendorPackPage
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/vendors/packs [get]
func (contr *VendorController) GetVendorShopPackPage(c *gin.Context) {
	rawPageNum := c.Query("pageNum")
	pageNum, err := strconv.ParseUint(rawPageNum, 10, 64)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}
	if pageNum <= 0 {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "page number must be positive"})
		return
	}
	searchStr := strings.ToLower(c.Query("search"))
	sortBy := c.Query("sortBy")
	sortDir := c.Query("sortDir")
	filterOn := c.Query("filterOn")

	vendorPage, err := contr.vendorService.GetVendorShopPackPage(
		c.Request.Context(),
		pageNum,
		sortBy,
		sortDir,
		filterOn,
		searchStr,
		c.Request.URL.String(),
		contr.categoryService,
	)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, vendorPage)
	return
}

// @Summary			Get vendor pack list page
// @Description		Get a list of packs specific to a vendor to see what they have in stock
// @Accept 			json
// @Produce			json
// @Param			uid path string true "vendor uid"
// @Param			pageNum query int true "page number"
// @Param			sortBy query string false "sort by"
// @Param			sortDir query string false "sort direction"
// @Param			filterOn query string false "category filter"
// @Param			search query string false "search string"
// @Tags			Vendor
// @Success			200 {object} model.VendorPackPage
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/vendor/packs/{uid} [get]
func (contr *VendorController) GetVendorPackListPage(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "valid uid param must be present",
		})
		return
	}
	rawPageNum := c.Query("pageNum")
	pageNum, err := strconv.ParseUint(rawPageNum, 10, 64)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}
	if pageNum <= 0 {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "page number must be positive"})
		return
	}
	searchStr := strings.ToLower(c.Query("search"))
	sortBy := c.Query("sortBy")
	sortDir := c.Query("sortDir")
	filterOn := c.Query("filterOn")

	vendorPage, err := contr.vendorService.GetVendorPackListPage(
		c.Request.Context(),
		authorizedUid,
		pageNum,
		sortBy,
		sortDir,
		filterOn,
		searchStr,
		c.Request.URL.String(),
		contr.categoryService,
	)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, vendorPage)
	return
}

// @Summary			Get vendor profile pack page
// @Description		Get a list of packs that a vendor has on their profile which users can see
// @Accept 			json
// @Produce			json
// @Param			uid path string true "vendor uid"
// @Param			pageNum query int true "page number"
// @Param			sortBy query string false "sort by"
// @Param			sortDir query string false "sort direction"
// @Param			filterOn query string false "category filter"
// @Tags			Vendor
// @Success			200 {object} model.VendorPackPage
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/vendor/packs/profile/{uid} [get]
func (contr *VendorController) GetVendorProfilePackPage(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "uid param must be set"})
		return
	}
	rawPageNum := c.Query("pageNum")
	pageNum, err := strconv.ParseUint(rawPageNum, 10, 64)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}
	if pageNum <= 0 {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "page number must be positive"})
		return
	}
	sortBy := c.Query("sortBy")
	sortDir := c.Query("sortDir")
	filterOn := c.Query("filterOn")

	vendorPage, err := contr.vendorService.GetVendorProfilePackPage(
		c.Request.Context(),
		uid,
		pageNum,
		sortBy,
		sortDir,
		filterOn,
		c.Request.URL.String(),
		contr.categoryService,
	)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, vendorPage)
	return
}

// @Summary			Get vendor item list page
// @Description		Get a list of item specific to a vendor to see what they have available to make packs
// @Accept 			json
// @Produce			json
// @Param			uid path string true "vendor uid"
// @Param			pageNum query int true "page number"
// @Param			sortBy query string false "sort by"
// @Param			sortDir query string false "sort direction"
// @Param			filterOn query string false "category filter"
// @Param			search query string false "search string"
// @Tags			Vendor
// @Success			200 {object} model.VendorItemPage
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/vendor/items/{uid} [get]
func (contr *VendorController) GetVendorItemListPage(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "valid uid param must be present",
		})
		return
	}
	rawPageNum := c.Query("pageNum")
	pageNum, err := strconv.ParseUint(rawPageNum, 10, 64)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}
	if pageNum <= 0 {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "page number must be positive"})
		return
	}
	searchStr := strings.ToLower(c.Query("search"))
	sortBy := c.Query("sortBy")
	sortDir := c.Query("sortDir")
	filterOn := c.Query("filterOn")

	vendorPage, err := contr.vendorService.GetVendorItemListPage(
		c.Request.Context(),
		authorizedUid,
		pageNum,
		sortBy,
		sortDir,
		filterOn,
		searchStr,
		c.Request.URL.String(),
		contr.categoryService,
		contr.itemService,
	)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, vendorPage)
	return
}

// @Summary 		Add vendor categories
// @Description 	Associates a set of categories to a vendor profile
// @Tags 			Vendor
// @Accept 			json
// @Produce 		json
// @Param 			categories body []model.VendorCategory true "categories"
// @Success 		200
// @Failure 		500 {object} httputil.HTTPError
// @Failure 		400 {object} httputil.HTTPError
// @Router 			/vendor/categories [post]
func (contr *VendorController) AddVendorCategories(c *gin.Context) {
	categories := []*model.VendorCategory{}
	if err := c.BindJSON(&categories); err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "valid uid param must be present",
		})
		return
	}

	for _, category := range categories {
		if category.VendorId != nil {
			if *category.VendorId != authorizedUid {
				httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
					Message: "user is not authorized to perform this action",
				})
				return
			}
		}
	}

	if len(categories) <= 0 {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "categories list is empty"})
		return
	}

	err := contr.vendorService.AddVendorCategories(c.Request.Context(), categories)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, "success")
	return
}

// @Summary 		Get vendor categories
// @Description 	Get a list of vendor categories
// @Tags 			Vendor
// @Accept 			json
// @Produce 		json
// @Param 			vendorId path string true "vendor uid"
// @Success 		200 {object} []model.VendorCategoryExpanded
// @Failure 		500 {object} httputil.HTTPError
// @Failure 		400 {object} httputil.HTTPError
// @Router 			/vendor/categories/:vendorId [get]
func (contr *VendorController) GetVendorCategories(c *gin.Context) {
	vendorId := c.Param("vendorId")
	if vendorId == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{
			Message: "vendorId query param must be present",
		})
		return
	}
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "valid uid param must be present",
		})
		return
	}
	if vendorId != authorizedUid {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "user is not authorized to perform this action",
		})
		return
	}

	categories, err := contr.vendorService.GetVendorCategories(c.Request.Context(), vendorId)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, categories)
	return
}

func (contr *VendorController) ClearVendorItemCache(c *gin.Context) {
	vendorId := c.Query("vendorId")
	if vendorId == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "vendorId query param must be present"})
		return
	}

	if err := contr.itemService.ClearVendorItemCache(c.Request.Context(), vendorId); err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, "success")
	return
}

func (contr *VendorController) ClearVendorPackCache(c *gin.Context) {
	vendorId := c.Query("vendorId")
	if vendorId == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "vendorId query param must be present"})
		return
	}

	if err := contr.packService.ClearVendorPackCache(c.Request.Context(), vendorId); err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, "success")
	return
}
