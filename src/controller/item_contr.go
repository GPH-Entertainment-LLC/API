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

type ItemController struct {
	itemService   service.ItemService
	vendorService service.VendorService
	packService   service.PackService
}

func NewItemController(itemService service.ItemService, vendorService service.VendorService, packService service.PackService) *ItemController {
	return &ItemController{itemService: itemService, vendorService: vendorService, packService: packService}
}

func (contr ItemController) Register(router *gin.Engine) {
	router.POST("/item", contr.CreateItem)
	router.POST("/item/categories", contr.AddItemCategories)
	router.GET("/item/:id", contr.GetItem)
	router.PATCH("/item/:id", contr.PatchItem)
	router.DELETE("/items/user", contr.DeleteUserItems)
	router.DELETE("/items", contr.DeleteItems)
}

type ErrorResponse struct {
	Errors []string `json:"errors"`
}

type EncodedImage struct {
	Data string `json:"data"`
}

// @Summary			Create an item
// @Description		Save item record in DB
// @Param			item body model.Item true "Item"
// @Accept			json
// @Produce			json
// @Tags			Item
// @Success			201 {object} model.Item
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/item [post]
func (contr ItemController) CreateItem(c *gin.Context) {
	item := model.Item{}
	if err := c.BindJSON(&item); err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}
	if item.VendorId == nil {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{
			Message: "Vendor ID cannot be null",
		})
		return
	}

	vendor, err := contr.vendorService.(service.VendorService).GetActiveVendor(c, *item.VendorId)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	authorizedUid := c.Query("authorizedUid")
	if *vendor.UID != authorizedUid {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "User is not authorized to perform this action",
		})
		return
	}

	createdItem, err := contr.itemService.(service.ItemService).CreateItem(c.Request.Context(), &item)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, createdItem)
	return
}

// @Summary			Add item categories
// @Description		Associate categories to an item for a vendor
// @Param			categories body []model.ItemCategory true "categories"
// @Accept 			json
// @Produce			json
// @Tags			Item
// @Success			200
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		400 {object} controller.ErrorResponse
// @Failure 		500 {object} httputil.HTTPError
// @Router			/item/categories [post]
func (contr ItemController) AddItemCategories(c *gin.Context) {
	categories := []*model.ItemCategory{}
	if err := c.BindJSON(&categories); err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	err := contr.itemService.AddItemCategories(c.Request.Context(), categories)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, "Success")
	return
}

// @Summary			Get an item
// @Description		Retrieve an item based on ID
// @Param			id path int true "Item ID"
// @Accept 			json
// @Produce			json
// @Tags			Item
// @Success			200 {object} model.Item
// @Failure 		500 {object} httputil.HTTPError
// @Router			/item/{id} [GET]
func (contr ItemController) GetItem(c *gin.Context) {
	rawItemId := c.Param("id")
	itemId, err := strconv.ParseUint(rawItemId, 10, 64)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}
	authorizedUid := c.Query("authorizedUid")

	item, err := contr.itemService.GetItem(c.Request.Context(), itemId)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	if *item.VendorId != authorizedUid {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "user is not authorized to perform this action",
		})
		return
	}

	c.JSON(http.StatusOK, item)
	return
}

// @Summary			Update an item
// @Description		Update an item with new patch
// @Param			itemPatch body model.Item true "item patch to update"
// @Param 			id path int true "item id"
// @Accept 			json
// @Produce			json
// @Tags			Item
// @Success			200 {object} model.Item
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/item/{id} [PATCH]
func (contr ItemController) PatchItem(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "authorized uid param must be present",
		})
		return
	}

	var itemPatch map[string]interface{}
	err := c.BindJSON(&itemPatch)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}
	rawItemId := c.Param("id")
	itemId, err := strconv.ParseUint(rawItemId, 10, 64)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
	}

	item, err := contr.itemService.PatchItem(c.Request.Context(), itemId, itemPatch, authorizedUid)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, item)
	return
}

// @Summary			Remove user items
// @Description		Deactivate user item records
// @Param 			ids query string true "user item ids"
// @Param			uid query string true "uid"
// @Accept 			json
// @Produce			json
// @Tags			Item
// @Success			200 {object} []int
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/items/user [DELETE]
func (contr ItemController) DeleteUserItems(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	rawUserItemIdsStr := c.Query("ids")
	rawUserItemIds := strings.Split(rawUserItemIdsStr, ",")
	userItemIds := make([]uint64, len(rawUserItemIds))

	for i, v := range rawUserItemIds {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			httputil.NewError(c, http.StatusBadRequest, err)
			return
		}
		userItemIds[i] = id
	}

	err := contr.itemService.DeleteUserItems(c.Request.Context(), userItemIds, authorizedUid)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, userItemIds)
	return
}

// @Summary			Delete an item
// @Description		Deactivate an item from the DB
// @Param 			ids query string true "item ids"
// @Param			vendorId query string true "vendor uid"
// @Accept 			json
// @Produce			json
// @Tags			Item
// @Success			200 {object} []int
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/items [DELETE]
func (contr ItemController) DeleteItems(c *gin.Context) {
	vendorId := c.Query("vendorId")
	if vendorId == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{
			Message: "vendorId param must be present",
		})
		return
	}

	authorizedUid := c.Query("authorizedUid")
	if vendorId != authorizedUid {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "user is not authorized to perform this action",
		})
		return
	}

	rawItemIdsStr := c.Query("ids")
	rawItemIds := strings.Split(rawItemIdsStr, ",")
	itemIds := make([]uint64, len(rawItemIds))
	for i, v := range rawItemIds {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			httputil.NewError(c, http.StatusBadRequest, err)
			return
		}
		itemIds[i] = id
	}

	if err := contr.itemService.DeleteItems(c.Request.Context(), itemIds, vendorId, contr.packService); err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, itemIds)
	return
}
