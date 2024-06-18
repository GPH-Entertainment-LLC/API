package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"xo-packs/core"
	"xo-packs/db"
	"xo-packs/model"
	"xo-packs/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/swaggo/swag/example/celler/httputil"
)

type UserController struct {
	userService   service.UserService
	vendorService service.VendorService
	itemService   service.ItemService
}

func NewUserController(userService service.UserService, vendorService service.VendorService, itemService service.ItemService) *UserController {
	return &UserController{userService: userService, vendorService: vendorService, itemService: itemService}
}

func (contr UserController) Register(router *gin.Engine) {
	router.POST("/user", contr.CreateUser)
	router.POST("/user/favorite", contr.AddFavorite)
	router.POST("/user/item/withdrawal", contr.WithdrawalUserItem)
	router.GET("/user/:uid", contr.GetUser)
	router.GET("/userExists/:uid", contr.GetUserByUid)
	router.GET("/user/username/:username", contr.GetUserByUsername)
	router.GET("/user/packs/:uid", contr.GetUserPackPage)
	router.GET("/user/items/:uid", contr.GetUserItemPage)
	router.GET("/user/favorite", contr.GetFavorite)
	router.GET("/user/favorites/:uid", contr.GetUserFavoritesPage)
	router.PATCH("/user", contr.PatchUser)
	router.DELETE("/user", contr.DeleteUser)
	router.DELETE("/user/clear", contr.ClearUserCache)
	router.DELETE("/user/favorite", contr.RemoveFavorite)
}

// @Summary			Create a user
// @Description		Save user record in DB
// @Param			user body model.User true "User"
// @Accept			json
// @Produce			json
// @Tags			User
// @Success			201 {object} model.User
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/user [post]
func (contr UserController) CreateUser(c *gin.Context) {
	user := model.User{}
	if err := c.BindJSON(&user); err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}
	if user.Uid == nil {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "uid must be present in new user creation"})
		return
	}
	if user.Username == nil {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "username must be present in new user creation"})
		return
	}
	if user.Email == nil {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "email must be present in new user creation"})
		return
	}

	authorizedUid := c.Query("authorizedUid")
	if *user.Uid != authorizedUid {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "user is not authorized to perform this action",
		})
		return
	}

	createdUser, err := contr.userService.CreateUser(c.Request.Context(), &user)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	// log new user
	core.AddLog(logrus.Fields{
		"UID":      *user.Uid,
		"Username": *user.Username,
		"Email":    *user.Email,
	}, c, db.LOG_USER_CREATE)

	c.JSON(http.StatusCreated, createdUser)
	return
}

// @Summary			Add a user favorite
// @Description		Add a vendor favorite record to a user
// @Param			uid query string true "uid"
// @Param			vendorId query string true "vendor uid"
// @Accept			json
// @Produce			json
// @Tags			User
// @Success			201 {object} model.Favorite
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/user/favorite [post]
func (contr UserController) AddFavorite(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "valid uid must be present"})
		return
	}

	vendorId := c.Query("vendorId")
	if vendorId == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "vendor id param must be present"})
		return
	}

	userFavorite, err := contr.userService.AddFavorite(c.Request.Context(), authorizedUid, vendorId, contr.vendorService)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, userFavorite)
	return
}

// @Summary			Withdrawal a user item
// @Description		Allows a user to withdrawal an item from their collection which is externally fulfilled
// @Param			uid query string true "uid"
// @Param			authorizedUid query string true "authorized uid"
// @Param			userItemId query int true "user item id"
// @Accept			json
// @Produce			json
// @Tags			User
// @Success			201 {object} model.ItemWithdrawal
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/user/item/withdrawal [post]
func (contr UserController) WithdrawalUserItem(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{
			Message: "authorizedUid must be present",
		})
		return
	}

	rawUserItemId := c.Query("userItemId")
	if rawUserItemId == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{
			Message: "user item id param must be present",
		})
	}
	userItemId, err := strconv.ParseUint(rawUserItemId, 10, 64)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	userItemWithdrawal, err := contr.userService.WithdrawalUserItem(c.Request.Context(), userItemId, authorizedUid)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, userItemWithdrawal)
	return
}

// @Summary			Add a user favorite
// @Description		Add a vendor favorite record to a user
// @Param			uid query string true "uid"
// @Param			vendorId query string true "vendor uid"
// @Accept			json
// @Produce			json
// @Tags			User
// @Success			201 {object} model.Favorite
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/user/favorite [delete]
func (contr UserController) RemoveFavorite(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "valid uid must be present"})
		return
	}

	vendorId := c.Query("vendorId")
	if vendorId == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "vendor id param must be present"})
		return
	}

	err := contr.userService.RemoveFavorite(c.Request.Context(), authorizedUid, vendorId, contr.vendorService)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, err)
	return
}

// @Summary			Get a user
// @Description		Save a user record in DB
// @Param			uid path string true "uid"
// @Accept			json
// @Produce			json
// @Tags			User
// @Success			200 {object} model.User
// @Failure 		500 {object} httputil.HTTPError
// @Router			/user/{uid} [get]
func (contr UserController) GetUser(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "authorized uid param must be present",
		})
		return
	}

	uid := c.Param("uid")
	if uid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{
			Message: "uid param must be present",
		})
		return
	}

	if authorizedUid != uid {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "user does not have access to this content",
		})
		return
	}

	user, err := contr.userService.GetUser(c.Request.Context(), authorizedUid, true)
	if err != nil {
		fmt.Println(err)
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, user)
	return
}

func (contr UserController) GetUserByUid(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{
			Message: "uid param must be present",
		})
		return
	}

	user, err := contr.userService.GetUserByUid(c.Request.Context(), uid)
	if err != nil {
		fmt.Println(err)
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	if user.Uid == nil {
		httputil.NewError(c, http.StatusNotFound, &core.ErrorResp{
			Message: "Could not find user by uid",
		})
		return
	}

	c.JSON(http.StatusOK, user)
	return
}

// @Summary			Get a user by username
// @Description		Retrieve a user given the username
// @Param			username path string true "username"
// @Accept			json
// @Produce			json
// @Tags			User
// @Success			200 {object} model.User
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/user/username/:username [get]
func (contr UserController) GetUserByUsername(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "username query param must be present"})
		return
	}

	user, err := contr.userService.GetUserByUsername(c.Request.Context(), username)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, user)
	return
}

// @Summary			Get user collection pack page
// @Description		Get a list of packs that a user has in their collections
// @Accept 			json
// @Produce			json
// @Param			uid path string true "vendor uid"
// @Param			pageNum query int true "page number"
// @Param			sortBy query string false "sort by"
// @Param			sortDir query string false "sort direction"
// @Param			filterOn query string false "category filter"
// @Param			search query string false "search string"
// @Tags			User
// @Success			200 {object} model.UserPackPage
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/user/packs/{uid} [get]
func (contr UserController) GetUserPackPage(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "valid uid param must be set"})
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

	userPackPage, err := contr.userService.GetUserPackPage(
		c.Request.Context(),
		authorizedUid,
		pageNum,
		sortBy,
		sortDir,
		filterOn,
		searchStr,
		c.Request.URL.String(),
	)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, userPackPage)
	return
}

// @Summary			Get user collection item page
// @Description		Get a list of items that a user has in their collections
// @Accept 			json
// @Produce			json
// @Param			uid path string true "vendor uid"
// @Param			pageNum query int true "page number"
// @Param			sortBy query string false "sort by"
// @Param			sortDir query string false "sort direction"
// @Param			filterOn query string false "category filter"
// @Param			search query string false "search string"
// @Tags			User
// @Success			200 {object} model.UserItemPage
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/user/items/{uid} [get]
func (contr UserController) GetUserItemPage(c *gin.Context) {
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

	userItemPage, err := contr.userService.GetUserItemPage(
		c.Request.Context(),
		authorizedUid,
		pageNum,
		sortBy,
		sortDir,
		filterOn,
		searchStr,
		c.Request.URL.String(),
		contr.itemService,
	)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, userItemPage)
	return
}

// @Summary			Get user favorite page
// @Description		Get user favorites vendor page
// @Accept 			json
// @Produce			json
// @Param			uid path string true "vendor uid"
// @Param			pageNum query int true "page number"
// @Param			search query string false "search string"
// @Tags			User
// @Success			200 {object} model.UserFavoritePage
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/user/favorites/{uid} [get]
func (contr UserController) GetUserFavoritesPage(c *gin.Context) {
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

	userFavoritesPage, err := contr.userService.GetUserFavoritesPage(c, authorizedUid, pageNum, searchStr, c.Request.URL.String())
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, userFavoritesPage)
	return
}

// @Summary			Get a favorite
// @Description		Get a favorite between user and vendor
// @Accept 			json
// @Produce			json
// @Param			uid query string true "uid"
// @Param			vendorId query string true "vendor uid"
// @Tags			User
// @Success			200 {object} model.Favorite
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/user/favorite [get]
func (contr UserController) GetFavorite(c *gin.Context) {
	uid := c.Query("uid")
	if uid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "uid query param must be present"})
		return
	}

	vendorId := c.Query("vendorId")
	if vendorId == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "vendorId query param must be present"})
		return
	}

	favorite, err := contr.userService.GetFavorite(c.Request.Context(), uid, vendorId)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, favorite)
	return
}

// @Summary			Patch a user
// @Description		Updates a user record in DB given a patch map
// @Param			uid query string true "uid"
// @Param			username query string true "username"
// @Param			userPatchMap body model.User true "user patch map"
// @Accept			json
// @Produce			json
// @Tags			User
// @Success			200 {object} model.User
// @Failure			400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/user [patch]
func (contr UserController) PatchUser(c *gin.Context) {
	var userPatchMap map[string]interface{}
	if err := c.BindJSON(&userPatchMap); err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "valid uid param must be present"})
		return
	}

	username := c.Query("username")
	if username == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "username query param must be present"})
		return
	}

	user, err := contr.userService.PatchUser(c.Request.Context(), authorizedUid, username, userPatchMap)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, user)
	return
}

// @Summary			Delete a user
// @Description		Deactivates a user from the DB
// @Param			uid query string true "uid"
// @Param			username query string true "username"
// @Accept			json
// @Produce			json
// @Tags			User
// @Success			200 {object} model.User
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/user [delete]
func (contr UserController) DeleteUser(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "valid uid must be present",
		})
		return
	}
	username := c.Query("username")
	if username == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "username query param must be present"})
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
	if user.Email == nil {
		httputil.NewError(c, http.StatusInternalServerError, &core.ErrorResp{Message: fmt.Sprintf("data quality error: user %v does not have email", authorizedUid)})
		return
	}
	if user.Username == nil {
		httputil.NewError(c, http.StatusInternalServerError, &core.ErrorResp{Message: fmt.Sprintf("data quality error: user %v does not have username", authorizedUid)})
		return
	}

	deletedUser, err := contr.userService.DeleteUser(c.Request.Context(), authorizedUid, username)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	// log user deletion
	core.AddLog(logrus.Fields{
		"UID":      authorizedUid,
		"Username": username,
		"Email":    *user.Email,
	}, c, db.LOG_USER_DELETE)

	c.JSON(http.StatusOK, deletedUser)
	return
}

// route requires admin privileges which uses different authentication than JWT
func (contr UserController) ClearUserCache(c *gin.Context) {
	uid := c.Query("uid")
	if uid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "valid uid must be present"})
		return
	}

	username := c.Query("username")
	if username == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "username param must be present"})
		return
	}

	if err := contr.userService.ClearUserCache(c.Request.Context(), uid, username); err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, "success")
	return
}
