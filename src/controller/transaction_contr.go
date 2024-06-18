package controller

import (
	"net/http"
	"strconv"
	"xo-packs/core"
	"xo-packs/model"
	"xo-packs/service"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/example/celler/httputil"
)

type TransactionController struct {
	transactionService service.TransactionService
	tokenService       service.TokenService
}

func NewTransactionController(transactionService service.TransactionService, tokenService service.TokenService) *TransactionController {
	return &TransactionController{transactionService: transactionService, tokenService: tokenService}
}

func (contr TransactionController) Register(router *gin.Engine) {
	router.POST("/transaction/newSale", contr.NewSale)
	router.POST("/transaction/charge", contr.ChargeTransaction)
	router.GET("/transaction/userInfo", contr.UserTransactionInfo)
	router.GET("/transaction/history/:uid", contr.GetUserTransactionHistoryPage)
}

// @Summary			Get a user transaction history page
// @Description		Get a list of trnasactions that a user has executed
// @Accept 			json
// @Produce			json
// @Param			uid path string true "uid"
// @Param			pageNum query int true "page number"
// @Tags			Transaction
// @Success			200 {object} model.UserTransactionHistoryPage
// @Failure 		400 {object} httputil.HTTPError
// @Failure 		500 {object} httputil.HTTPError
// @Router			/transaction/history/{uid} [get]
func (contr TransactionController) GetUserTransactionHistoryPage(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")

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

	transactionHistory, err := contr.transactionService.GetUserTransactionHistoryPage(c.Request.Context(), authorizedUid, pageNum)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	if transactionHistory == nil {
		httputil.NewError(c, http.StatusNotFound, &core.ErrorResp{
			Message: "ERROR: could not pull user transaction history at this time. Please contact support@ccbill.com for assitance",
		})
		return
	}

	c.JSON(http.StatusOK, transactionHistory)
	return
}

func (contr TransactionController) UserTransactionInfo(c *gin.Context) {
	authorizedUid := c.Query("authorizedUid")
	uid := c.Query("uid")

	if authorizedUid != uid {
		httputil.NewError(c, http.StatusUnauthorized, &core.ErrorResp{
			Message: "user is unauthorized to perform this action",
		})
	}

	userTransactionInfo, err := contr.transactionService.GetUserTransactionInfo(c.Request.Context(), uid)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	if userTransactionInfo == nil {
		httputil.NewError(c, http.StatusNotFound, &core.ErrorResp{
			Message: "user does not have any previous transactions",
		})
		return
	}

	c.JSON(http.StatusOK, userTransactionInfo)
	return
}

func (contr TransactionController) NewSale(c *gin.Context) {
	newSalesTransaction := model.NewSalesTransaction{}
	if err := c.BindJSON(&newSalesTransaction); err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	completedTransaction, err := contr.transactionService.NewSale(c.Request.Context(), &newSalesTransaction, contr.tokenService)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, completedTransaction)
	return
}

func (contr TransactionController) ChargeTransaction(c *gin.Context) {
	transaction := model.Transaction{}
	if err := c.BindJSON(&transaction); err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	completedTransaction, err := contr.transactionService.ChargeTransaction(c.Request.Context(), &transaction, contr.tokenService)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, completedTransaction)
	return
}
