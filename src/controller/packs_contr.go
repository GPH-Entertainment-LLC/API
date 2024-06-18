package controller

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
	"xo-packs/core"
	"xo-packs/db"
	"xo-packs/model"
	"xo-packs/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/swaggo/swag/example/celler/httputil"
)

type PackController struct {
	packService   service.PackService
	vendorService service.VendorService
	itemService   service.ItemService
	userService   service.UserService
	tokenService  service.TokenService
}

func NewPackController(
	packService service.PackService,
	vendorService service.VendorService,
	itemService service.ItemService,
	userService service.UserService,
	tokenService service.TokenService,
) *PackController {
	return &PackController{
		packService:   packService,
		vendorService: vendorService,
		itemService:   itemService,
		userService:   userService,
		tokenService:  tokenService,
	}
}

func (contr PackController) Register(router *gin.Engine) {
	router.POST("/pack/config", contr.CreatePackConfig)
	router.POST("/pack/item/configs", contr.AddPackItemConfigs)
	router.POST("/pack/generate", contr.GeneratePacks)
	router.POST("/pack/buy", contr.BuyPacks) // associates packs to user
	router.POST("/pack/categories", contr.AddPackCategories)
	router.GET("/pack/config/:id", contr.GetPackConfig)
	router.GET("/pack/open/:id", contr.OpenPack) // associates pack items to user and returns Pack obj
	router.GET("/packs/amount/:uid", contr.GetUserPackAmount)
	router.GET("/pack/items/:id", contr.GetPackItems)
	router.GET("/pack/items/preview/:id", contr.GetPackItemsPreview)
	router.POST("/pack/items/generateOdds", contr.GeneratePackItemOdds)
	router.POST("/pack/activate", contr.ActivatePacks)
	// router.POST("/pack/schedule", contr.SchedulePacks)
	router.PATCH("/pack/config", contr.PatchPackConfig)
	router.DELETE("/pack/deactivate", contr.DeactivatePacks)
	router.DELETE("/pack/configs", contr.DeletePackConfigs)
	router.DELETE("/packs/clearCache", contr.ClearPackShopCache)
}

// @Summary			Create a pack config
// @Description		Save pack configuration record in DB
// @Accept 			json
// @Produce			json
// @Param			pack body model.PackConfig true "Pack Config"
// @Tags			Pack
// @Success			200 {object} model.PackConfig
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/pack/config [post]
func (contr PackController) CreatePackConfig(c *gin.Context) {
	pack := model.PackConfig{}
	if err := c.BindJSON(&pack); err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}
	if pack.VendorID == nil {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "Pack must have a vendor ID associated"})
		return
	}

	vendor, err := contr.vendorService.GetActiveVendor(c, *pack.VendorID)
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

	createdPack, err := contr.packService.CreatePackConfig(c.Request.Context(), &pack, contr.vendorService)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, createdPack)
	return
}

// @Summary 		get packs
// @Description 	get a pack by id
// @Tags 			Pack
// @Accept 			json
// @Produce 		json
// @Param 			id path int true "Pack Config ID"
// @Success 		200 {object} model.PackConfig
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router 			/pack/config/{id} [get]
func (contr PackController) GetPackConfig(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	packConfig, err := contr.packService.GetPackConfig(c.Request.Context(), id)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	authorizedUid := c.Query("authorizedUid")
	if *packConfig.VendorID != authorizedUid {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "user is not authorized to perform this action",
		})
		return
	}

	c.JSON(http.StatusCreated, packConfig)
	return
}

// @Summary 		Get user pack amount
// @Description 	Get a user packs amount
// @Tags 			Pack
// @Accept 			json
// @Produce 		json
// @Success 		200 {object} core.AmountResp
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router 			/packs/amount/{uid} [get]
func (contr PackController) GetUserPackAmount(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "uid param must be present"})
		return
	}

	amount, err := contr.packService.GetUserPackAmount(c.Request.Context(), uid)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, core.AmountResp{Amount: *amount})
}

// @Summary 		Get a pack
// @Description 	Get a joined 2 dimensional pack object with nested item list
// @Tags 			Pack
// @Accept 			json
// @Produce 		json
// @Param 			id path int true "pack id"
// @Success 		200 {object} model.Pack
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router 			/pack/{id} [get]
func (contr PackController) GetPack(c *gin.Context) {
	rawId := c.Param("id")
	id, err := strconv.ParseUint(rawId, 10, 64)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	pack, err := contr.packService.GetPack(c.Request.Context(), id)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, pack)
	return
}

// @Summary 		Get pack item configs
// @Description 	Get a list of pack item configs
// @Tags 			Pack
// @Accept 			json
// @Produce 		json
// @Param 			id path int true "pack id"
// @Success 		200 {object} []model.PackItemConfigExpanded
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router 			/pack/items/:id [get]
func (contr PackController) GetPackItems(c *gin.Context) {
	rawId := c.Param("id")
	id, err := strconv.ParseUint(rawId, 10, 64)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}
	authorizedUid := c.Query("authorizedUid")

	packConfig, err := contr.packService.GetPackConfig(c.Request.Context(), id)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	if *packConfig.VendorID != authorizedUid {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "user is not authorized to perform this action",
		})
		return
	}

	packItems, err := contr.packService.GetPackItems(c.Request.Context(), id, contr.itemService)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, packItems)
	return
}

// should be the same as get pack items but add logic for checking external fulfillment and subtituting preview image
// @Summary 		Get pack item configs preview
// @Description 	Get a list of pack item configs preview
// @Param			id path int true "pack config id"
// @Tags 			Pack
// @Accept 			json
// @Produce 		json
// @Param 			id path int true "pack id"
// @Success 		200 {object} []model.PackItemConfigExpanded
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router 			/pack/items/preview/:id [get]
func (contr PackController) GetPackItemsPreview(c *gin.Context) {
	rawId := c.Param("id")
	id, err := strconv.ParseUint(rawId, 10, 64)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	previewPackItems, err := contr.packService.GetPackItemsPreview2(c.Request.Context(), id)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, previewPackItems)
	return
}

// @Summary 		Open a pack
// @Description 	Sets an opened date on the pack fact
// @Tags 			Pack
// @Accept 			json
// @Produce 		json
// @Param 			id path int true "pack fact id"
// @Param			uid query string true "uid"
// @Success 		200 {object} model.Pack
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router 			/pack/open/{id} [get]
func (contr PackController) OpenPack(c *gin.Context) {
	rawId := c.Param("id")
	id, err := strconv.ParseUint(rawId, 10, 64)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}
	authorizedUid := c.Query("authorizedUid")

	pack, err := contr.packService.GetPack(c.Request.Context(), id)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	if *pack.OwnerId != authorizedUid {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "user is not authorized to perform this action",
		})
		return
	}

	pack, err = contr.packService.OpenPack(c.Request.Context(), id, authorizedUid, contr.itemService)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	if pack.ID == nil || len(pack.Items) <= 0 {
		httputil.NewError(c, http.StatusInternalServerError, &core.ErrorResp{
			Message: fmt.Sprintf("Critical error: For some reason we were unable to get data related to this pack. Please contact customer-support@xopacks.com for help"),
		})
		return
	}

	if pack.VendorId == nil {
		httputil.NewError(c, http.StatusInternalServerError, &core.ErrorResp{
			Message: fmt.Sprintf("Critical error: pack id %v did not have a creator associated. Please contact customer-support@xopacks.com to get this resolved", *pack.ID),
		})
		return
	}

	user, err := contr.userService.GetUser(c.Request.Context(), authorizedUid, true)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, &core.ErrorResp{
			Message: "An error occurred getting the user meta data. Please contact customer-support@xopacks.com to resolve this issue",
		})
		return
	}
	if user.Username == nil || user.Email == nil {
		httputil.NewError(c, http.StatusInternalServerError, &core.ErrorResp{
			Message: "Data quality error with user meta data. Missing uid, username, or email. Please contact customer-support@xopacks.com to resolve this issue",
		})
		return
	}

	vendor, err := contr.vendorService.GetVendor(c.Request.Context(), *pack.VendorId)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	applyNotificationLog := true
	if vendor.UID == nil {
		fmt.Println("Creator does not exist")
		applyNotificationLog = false
	}
	if vendor.Email == nil ||
		vendor.Username == nil ||
		vendor.FirstName == nil ||
		vendor.LastName == nil {
		fmt.Println("data quality error: creator does not have all necessary information")
		fmt.Println("==== VENDOR UID: ", vendor.UID)
		applyNotificationLog = false
	}

	// logging pack open
	packOpenLog := logrus.Fields{
		"UID":       authorizedUid,
		"Username":  *user.Username,
		"UserEmail": *user.Email,
		"PackId":    *pack.ID,
	}
	core.AddLog(packOpenLog, c, db.LOG_PACK_OPEN)

	// logging item pulls
	if applyNotificationLog {
		numWorkers := len(pack.Items)
		for i := 0; i < numWorkers; i++ {
			go func(idx int) {
				if pack.Items[idx].Notify != nil && *pack.Items[idx].Notify {
					core.AddLog(logrus.Fields{
						"UID":              authorizedUid,
						"Username":         *user.Username,
						"UserEmail":        *user.Email,
						"PackId":           *pack.ID,
						"ItemId":           *pack.Items[idx].ID,
						"ItemName":         *pack.Items[idx].Name,
						"CreatorId":        *vendor.UID,
						"CreatorEmail":     *vendor.Email,
						"CreatorFirstName": *vendor.FirstName,
						"CreatorLastName":  *vendor.LastName,
					}, c, db.LOG_ITEM_NOTIFY_PULLS)
				} else {
					core.AddLog(logrus.Fields{
						"UID":              authorizedUid,
						"PackId":           *pack.ID,
						"ItemId":           *pack.Items[idx].ID,
						"CreatorId":        *vendor.UID,
						"CreatorEmail":     *vendor.Email,
						"CreatorFirstName": *vendor.FirstName,
						"CreatorLastName":  *vendor.LastName,
					}, c, db.LOG_ITEM_PULLS)
				}
			}(i)
		}
	}

	c.JSON(http.StatusOK, pack)
	return
}

// @Summary 		Add pack item configs
// @Description 	Add pack item configs that determine how packs are built
// @Tags 			Pack
// @Accept 			json
// @Produce 		json
// @Param 			packItemConfigs body []model.PackItemConfig true "Pack Item Configs"
// @Success 		201 {object} core.AddPackSuccessResp
// @Failure 		500 {object} httputil.HTTPError
// @Failure 		400 {object} httputil.HTTPError
// @Router 			/pack/item/configs [post]
func (contr PackController) AddPackItemConfigs(c *gin.Context) {
	var packItemConfigs []*model.PackItemConfig
	if err := c.BindJSON(&packItemConfigs); err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}
	authorizedUid := c.Query("authorizedUid")

	for _, itemConfig := range packItemConfigs {
		packConfig, err := contr.packService.GetPackConfig(c.Request.Context(), *itemConfig.PackConfigID)
		if err != nil {
			if *packConfig.VendorID != authorizedUid {
				httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
					Message: "user is not authorized to perform this action",
				})
				return
			}
		}
	}

	if len(packItemConfigs) <= 0 {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "pack item configs list must not be empty"})
		return
	}

	if err := contr.packService.AddPackItemConfigs(c.Request.Context(), packItemConfigs); err != nil {
		httputil.NewError(c, http.StatusBadGateway, err)
		return
	}

	c.JSON(http.StatusCreated, core.AddPackSuccessResp{Message: "Successfully added the pack item configurations"})
	return
}

// @Summary 		Generate pack stock
// @Description 	Add pack and item fact records to DB for logical pack stock quantities
// @Tags 			Pack
// @Accept 			json
// @Produce 		json
// @Param 			packConfigId query int true "Pack Config ID"
// @Param			vendorId query string true "vendor uid"
// @Success 		201 {object} core.AddPackSuccessResp
// @Failure 		500 {object} httputil.HTTPError
// @Failure 		400 {object} httputil.HTTPError
// @Router 			/pack/generate [post]
func (contr PackController) GeneratePacks(c *gin.Context) {
	rawPackConfigId := c.Query("packConfigId")
	packConfigId, err := strconv.ParseUint(rawPackConfigId, 10, 64)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	vendorId := c.Query("vendorId")
	if vendorId == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "vendor uid must be present in params"})
		return
	}

	authorizedUid := c.Query("authorizedUid")
	if vendorId != authorizedUid {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "user is not authorized to perform this action",
		})
		return
	}

	if err := contr.packService.GeneratePacks(c.Request.Context(), packConfigId, vendorId, contr.vendorService); err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, core.AddPackSuccessResp{Message: "Successfully added bulk stock of packs"})
	return
}

// @Summary 		Buys a set of packs
// @Description 	Associates a new set of x available packs with a user after purchasing
// @Tags 			Pack
// @Accept 			json
// @Produce 		json
// @Param 			packConfigId query int true "Pack Config ID"
// @Param 			uid query string true "uid"
// @Param 			amount query int true "amount of packs"
// @Success 		201 {object} model.PackBoughtResp
// @Failure 		500 {object} httputil.HTTPError
// @Failure 		400 {object} httputil.HTTPError
// @Router 			/pack/buy [post]
func (contr PackController) BuyPacks(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	rawPackConfigId := c.Query("packConfigId")
	packConfigId, err := strconv.ParseUint(rawPackConfigId, 10, 64)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	rawAmount := c.Query("amount")
	amount, err := strconv.ParseInt(rawAmount, 10, 64)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	if authorizedUid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "Must supply a valid uid query param"})
		return
	}

	var resp *model.PackBoughtResp
	attempt := 0
	for attempt < 3 {
		resp, err = contr.packService.BuyPacks(c.Request.Context(), authorizedUid, packConfigId, float64(amount), contr.tokenService)
		if err != nil {
			if reflect.TypeOf(err) != reflect.TypeOf(core.DBErrorResp{}) {
				break
			}
		} else {
			break
		}
		attempt += 1
		time.Sleep(time.Second * 2) // speeing before next call
	}
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, resp)
	return
}

// @Summary 		Add pack categories
// @Description 	Associates a set of categories to a pack config
// @Tags 			Pack
// @Accept 			json
// @Produce 		json
// @Param 			categories body []model.PackCategory true "categories"
// @Success 		200
// @Failure 		500 {object} httputil.HTTPError
// @Failure 		400 {object} httputil.HTTPError
// @Router 			/packs/categories [post]
func (contr PackController) AddPackCategories(c *gin.Context) {
	categories := []*model.PackCategory{}
	err := c.BindJSON(&categories)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}
	authorizedUid := c.Query("authorizedUid")

	for _, category := range categories {
		packConfig, err := contr.packService.GetPackConfig(c.Request.Context(), *category.PackConfigId)
		if err != nil {
			httputil.NewError(c, http.StatusInternalServerError, err)
			return
		}
		if *packConfig.VendorID != authorizedUid {
			httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
				Message: "user is not authorized to perform this action",
			})
			return
		}
	}

	err = contr.packService.AddPackCategories(c.Request.Context(), categories)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, "Success")
	return
}

// @Summary 		Patch a pack config
// @Description 	Update a pack config
// @Tags 			Pack
// @Accept 			json
// @Produce 		json
// @Param			packConfigId query int true "pack config id"
// @Param			vendorId query string true "vendor uid"
// @Param			patchMap body model.PackConfig true "pack config patch map"
// @Success 		200
// @Failure 		500 {object} httputil.HTTPError
// @Failure 		400 {object} httputil.HTTPError
// @Router 			/pack/config [patch]
func (contr PackController) PatchPackConfig(c *gin.Context) {
	vendorId := c.Query("vendorId")
	if vendorId == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{
			Message: "vendorId query param must be present",
		})
	}

	authorizedUid := c.Query("authorizedUid")
	if authorizedUid != vendorId {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "user is not authorized to perform this action",
		})
		return
	}

	rawPackConfigId := c.Query("packConfigId")
	packConfigId, err := strconv.ParseUint(rawPackConfigId, 10, 64)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	patchMap := map[string]interface{}{}
	if err = c.BindJSON(&patchMap); err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	packConfig, err := contr.packService.PatchPackConfig(c.Request.Context(), packConfigId, patchMap, vendorId)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, packConfig)
	return
}

func (contr PackController) ActivatePacks(c *gin.Context) {
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
	}

	rawPackConfigIdsStr := c.Query("ids")
	rawPackConfigIds := strings.Split(rawPackConfigIdsStr, ",")
	packConfigIds := make([]uint64, len(rawPackConfigIds))
	for i, v := range rawPackConfigIds {
		packConfigId, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			httputil.NewError(c, http.StatusBadRequest, err)
			return
		}
		packConfigIds[i] = packConfigId
	}

	if len(packConfigIds) <= 0 {
		c.JSON(http.StatusOK, nil)
		return
	}

	err := contr.packService.ActivatePacks(c.Request.Context(), packConfigIds, vendorId, contr.vendorService)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, packConfigIds)
	return
}

// @Summary 		Inactivate pack(s) from the marketplace
// @Description 	Set an active false value without soft deleting a pack to make pack private without release date to customers
// @Tags 			Pack
// @Accept 			json
// @Produce 		json
// @Param 			ids query string true "pack config ids"
// @Param			vendorId query string true "vendor id"
// @Success 		200
// @Failure 		500 {object} httputil.HTTPError
// @Failure 		400 {object} httputil.HTTPError
// @Router 			/pack/deactivate [delete]
func (contr PackController) DeactivatePacks(c *gin.Context) {
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
	}

	rawPackConfigIdsStr := c.Query("ids")
	rawPackConfigIds := strings.Split(rawPackConfigIdsStr, ",")
	packConfigIds := make([]uint64, len(rawPackConfigIds))
	for i, v := range rawPackConfigIds {
		packConfigId, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			httputil.NewError(c, http.StatusBadRequest, err)
			return
		}
		packConfigIds[i] = packConfigId
	}

	if len(packConfigIds) <= 0 {
		c.JSON(http.StatusOK, nil)
		return
	}

	err := contr.packService.DeactivatePacks(c.Request.Context(), packConfigIds, vendorId, contr.vendorService)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, packConfigIds)
	return
}

// @Summary 		Delete a list of pack config
// @Description 	Soft delete pack configs and all item pack configs in DB
// @Tags 			Pack
// @Accept 			json
// @Produce 		json
// @Param 			ids query string true "pack config ids"
// @Param			vendorId query string true "vendor id"
// @Success 		200
// @Failure 		500 {object} httputil.HTTPError
// @Failure 		400 {object} httputil.HTTPError
// @Router 			/pack/configs [delete]
func (contr PackController) DeletePackConfigs(c *gin.Context) {
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
	}

	rawPackConfigIdsStr := c.Query("ids")
	rawPackConfigIds := strings.Split(rawPackConfigIdsStr, ",")
	packConfigIds := make([]uint64, len(rawPackConfigIds))
	for i, v := range rawPackConfigIds {
		packConfigId, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			httputil.NewError(c, http.StatusBadRequest, err)
			return
		}
		packConfigIds[i] = packConfigId
	}

	if len(packConfigIds) <= 0 {
		c.JSON(http.StatusOK, nil)
		return
	}

	err := contr.packService.DeletePackConfigs(c.Request.Context(), packConfigIds, vendorId, contr.vendorService)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, packConfigIds)
	return
}

// @Summary 		Generate odds to use for pack item configurations
// @Description 	Given a list of items to distribute in a pack, return odds distribution based on item rarities
// @Tags 			Pack
// @Accept 			json
// @Produce 		json
// @Param			authorizedUid query string true "authorized uid"
// @Param			vendorId query string true "vendor id"
// @Success 		200 {object} []int
// @Failure 		500 {object} httputil.HTTPError
// @Failure 		400 {object} httputil.HTTPError
// @Router 			/pack/items/generateOdds [POST]
func (contr PackController) GeneratePackItemOdds(c *gin.Context) {
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

	rawTotalItems := c.Query("totalItems")
	if rawTotalItems == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{
			Message: "total items param cannot be empty",
		})
		return
	}
	totalItems, err := strconv.ParseInt(rawTotalItems, 10, 64)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	items := []model.PackItemConfig{}
	if err := c.BindJSON(&items); err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
	}

	result, err := contr.packService.GeneratePackItemOdds(c.Request.Context(), vendorId, items, int(totalItems), contr.itemService)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, result)
	return
}

func (contr PackController) ClearPackShopCache(c *gin.Context) {
	if err := contr.packService.ClearPackShopCache(c.Request.Context()); err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, "success")
	return
}
