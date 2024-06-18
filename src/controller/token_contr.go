package controller

import (
	"net/http"
	"strconv"
	"xo-packs/core"
	"xo-packs/service"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/example/celler/httputil"
)

type TokenController struct {
	tokenService service.TokenService
}

func NewTokenController(tokenService service.TokenService) *TokenController {
	return &TokenController{tokenService: tokenService}
}

func (contr TokenController) Register(router *gin.Engine) {
	router.POST("/token/bundle", contr.AddBundle)
	router.POST("/token/buy/:uid", contr.BuyTokens)
	router.GET("/token/currencyRate", contr.ActiveTokenRate)
	router.GET("/token/bundle/:id", contr.GetBundle)
	router.GET("/token/bundles", contr.GetCurrentBundles)
	router.GET("/token/bundles/price", contr.GetBundlesByPrice)
	router.GET("/token/bundles/priceRange", contr.GetBundlesByPriceRange)
	router.GET("/token/balance/:uid", contr.GetBalance)
	router.PATCH("/token/balance/:uid", contr.UpdateBalance)
}

// @Summary 		Add a token bundle
// @Description 	Add a token bundle to the DB
// @Tags 			Token
// @Accept 			json
// @Produce 		json
// @Param 			dollarAmt query number true "dollar amount"
// @Param 			tokenAmt query number true "token amount"
// @Param			bundleImageId query int true "bundle image id"
// @Success 		200 {object} model.TokenBundle
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router 			/token/bundle [post]
func (contr *TokenController) AddBundle(c *gin.Context) {
	rawDollarAmt := c.Query("dollarAmt")
	rawTokenAmt := c.Query("tokenAmt")
	bundleImageUrl := c.Query("bundleImageUrl")

	dollarAmt, err := strconv.ParseFloat(rawDollarAmt, 64)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	tokenAmt, err := strconv.ParseFloat(rawTokenAmt, 64)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	tokenBundle, err := contr.tokenService.AddBundle(c.Request.Context(), &dollarAmt, &tokenAmt, bundleImageUrl)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, tokenBundle)
	return
}

func (contr *TokenController) BuyTokens(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	rawAmount := c.Query("amount")
	amount, err := strconv.ParseUint(rawAmount, 10, 64)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	txn, err := contr.tokenService.BuyTokens(c.Request.Context(), authorizedUid, amount, "12345")
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, txn)
	return
}

func (contr *TokenController) ActiveTokenRate(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "authorized uid must be present",
		})
		return
	}

	activeRate, err := contr.tokenService.ActiveTokenRate(c)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, activeRate)
	return
}

// @Summary 		Get a token bundle
// @Description 	Get a token bundle by id
// @Tags 			Token
// @Accept 			json
// @Produce 		json
// @Param 			id path int true "bundle id"
// @Success 		200 {object} model.TokenBundle
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router 			/token/bundle/{id} [get]
func (contr *TokenController) GetBundle(c *gin.Context) {
	rawId := c.Param("id")
	id, err := strconv.ParseUint(rawId, 10, 64)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	tokenBundle, err := contr.tokenService.GetBundle(c.Request.Context(), id)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, tokenBundle)
	return
}

// @Summary 		Get a token balance
// @Description 	Get a token balance for a user
// @Tags 			Token
// @Accept 			json
// @Produce 		json
// @Param 			id path int true "bundle id"
// @Success 		200 {object} model.TokenBalance
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router 			/token/balance/{uid} [get]
func (contr *TokenController) GetBalance(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	if authorizedUid == "" {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "valid uid must be present",
		})
		return
	}

	tokenBalance, err := contr.tokenService.GetBalance(c.Request.Context(), authorizedUid)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, tokenBalance)
	return
}

// @Summary 		Get token bundles
// @Description 	Get a list of current token bundles
// @Tags 			Token
// @Accept 			json
// @Produce 		json
// @Success 		200 {object} []model.TokenBundle
// @Failure 		500 {object} httputil.HTTPError
// @Router 			/token/bundles [get]
func (contr *TokenController) GetCurrentBundles(c *gin.Context) {
	tokenBundles, err := contr.tokenService.GetCurrentBundles(c.Request.Context())
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, tokenBundles)
	return
}

// @Summary 		Get a token bundles filtered by dollar amount
// @Description 	Get a token bundles filtered by dollar amount
// @Tags 			Token
// @Accept 			json
// @Produce 		json
// @Param 			dollarAmt query number true "dollar amount"
// @Success 		200 {object} model.TokenBalance
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router 			/token/bundles/price [get]
func (contr *TokenController) GetBundlesByPrice(c *gin.Context) {
	rawDollarAmt := c.Query("dollarAmt")
	dollarAmt, err := strconv.ParseFloat(rawDollarAmt, 64)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	tokenBundles, err := contr.tokenService.GetBundlesByPrice(c.Request.Context(), &dollarAmt)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, tokenBundles)
	return
}

// @Summary 		Get a list of token bundles filtered on lower upper limits
// @Description 	Get a list of token bundles filtered on lower upper limits
// @Tags 			Token
// @Accept 			json
// @Produce 		json
// @Param 			lowerDollarAmt query number true "lower dollar amount"
// @Param 			upperDollarAmt query number true "upper dollar amount"
// @Success 		200 {object} []model.TokenBundle
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router 			/token/bundles/priceRange [get]
func (contr *TokenController) GetBundlesByPriceRange(c *gin.Context) {
	rawLowerDollarAmt := c.Query("lowerDollarAmt")
	lowerDollarAmt, err := strconv.ParseFloat(rawLowerDollarAmt, 64)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	rawUpperDollarAmt := c.Query("upperDollarAmt")
	upperDollarAmt, err := strconv.ParseFloat(rawUpperDollarAmt, 64)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	tokenBundles, err := contr.tokenService.GetBundlesByPriceRange(c.Request.Context(), &lowerDollarAmt, &upperDollarAmt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, tokenBundles)
	return
}

// @Summary 		Update a token balance
// @Description 	Update a users token balance
// @Tags 			Token
// @Accept 			json
// @Produce 		json
// @Param 			uid path int true "uid"
// @Success 		200 {object} model.TokenBalance
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router 			/token/balance/{uid} [patch]
func (contr *TokenController) UpdateBalance(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		httputil.NewError(c, http.StatusBadRequest, &core.ErrorResp{Message: "uid path param must be present"})
		return
	}
	var patchMap map[string]interface{}
	if err := c.BindJSON(&patchMap); err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	tokenBalance, err := contr.tokenService.UpdateBalance(c.Request.Context(), uid, patchMap)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, tokenBalance)
	return
}
