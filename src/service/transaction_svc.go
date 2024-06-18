package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"xo-packs/core"
	"xo-packs/model"
	"xo-packs/repository"
)

type TransactionService interface {
	NewSale(context.Context, *model.NewSalesTransaction, TokenService) (*model.NewSalesTransaction, error)
	GetUserTransactionInfo(context.Context, string) (*model.UserTransactionInfo, error)
	ChargeTransaction(context.Context, *model.Transaction, TokenService) (*model.Transaction, error)
	GetUserTransactionHistoryPage(context.Context, string, uint64) (*model.UserTransactionHistoryPage, error)
	GetCharge(context.Context)
}

type TransactionSvcImpl struct {
	transactionRepo repository.TransactionRepository
}

func NewTransactionService(repo repository.TransactionRepository) TransactionService {
	return &TransactionSvcImpl{transactionRepo: repo}
}

func (service *TransactionSvcImpl) GetUserTransactionInfo(c context.Context, uid string) (*model.UserTransactionInfo, error) {
	return service.transactionRepo.GetUserTransactionInfo(c, uid)
}

func (service *TransactionSvcImpl) GetUserTransactionHistoryPage(c context.Context, uid string, pageNumber uint64) (*model.UserTransactionHistoryPage, error) {
	adjustedPageNum := pageNumber - 1
	userTransactionHistoryPage, transactionAmount, err := service.transactionRepo.GetUserTransactionHistoryPage(c, uid, adjustedPageNum)
	if err != nil {
		return nil, err
	}
	pageSizeStr := os.Getenv("TRANSACTION_HISTORY_PAGE_SIZE")
	pageSize, err := strconv.ParseUint(pageSizeStr, 10, 64)
	if err != nil {
		return nil, err
	}
	thisPageSize := uint64(len(userTransactionHistoryPage))
	nextPageNum := uint64(0)
	nextPage := &nextPageNum
	if thisPageSize < pageSize {
		nextPage = nil
	} else {
		nextPageNum = (pageNumber + uint64(1))
	}
	result := model.UserTransactionHistoryPage{TransactionAmount: transactionAmount, PageSize: &thisPageSize, NextPage: nextPage, Page: userTransactionHistoryPage}
	return &result, nil
}

func (service *TransactionSvcImpl) NewSale(c context.Context, newSaleTxn *model.NewSalesTransaction, tokenService TokenService) (*model.NewSalesTransaction, error) {
	// null checking on required fields
	if newSaleTxn.Uid == nil || newSaleTxn.TransactionId == nil {
		return nil, &core.ErrorResp{
			Message: "ERROR: transaction data must be present",
		}
	}

	// getting the token bundle ID lookup using transaction amount
	tokenBundle, err := tokenService.GetBundleByPrice(c, newSaleTxn.AccountingInitialPrice)
	if err != nil {
		return nil, err
	}
	if tokenBundle.ID == nil {
		return nil, &core.ErrorResp{
			Message: "ERROR: no token bundle exists with this accounting initial price",
		}
	}
	newSaleTxn.TokenBundleId = tokenBundle.ID
	fmt.Println("Transaction ID: ", *newSaleTxn.TransactionId)

	// upload the new sale transaction
	completedNewSaleTxn, err := service.transactionRepo.NewSale(c, newSaleTxn)
	if err != nil {
		return nil, err
	}
	fmt.Println("Completed New Sale Transaction: ", completedNewSaleTxn)

	// create transaction model
	txn := model.Transaction{
		Uid:             *newSaleTxn.Uid,
		TokenBundleId:   int(*tokenBundle.ID),
		TokenAmount:     *tokenBundle.TokenAmount,
		TransactionId:   *newSaleTxn.TransactionId,
		SubscriptionId:  fmt.Sprintf("%v", *newSaleTxn.SubscriptionId),
		TranDatetime:    *newSaleTxn.TranDatetime,
		ClientAccNum:    *newSaleTxn.ClientAccNum,
		ClientSubAcc:    *newSaleTxn.ClientSubAcc,
		InitialPrice:    *newSaleTxn.BilledInitialPrice,
		InitialPeriod:   fmt.Sprintf("%v", *newSaleTxn.InitialPeriod),
		RecurringPrice:  *newSaleTxn.BilledRecurringPrice,
		RecurringPeriod: fmt.Sprintf("%v", *newSaleTxn.RecurringPeriod),
		Rebills:         fmt.Sprintf("%v", *newSaleTxn.Rebills),
		CurrencyCode:    fmt.Sprintf("%v", *newSaleTxn.CurrencyCode),
	}

	// charge transaction
	completedTxn, err := service.transactionRepo.ChargeTransaction(c, &txn)
	if err != nil {
		return nil, err
	}
	fmt.Println("Completed Transaction: ", completedTxn)

	// buy the tokens and update user token balance
	updatedBalance, err := tokenService.BuyTokens(c, *newSaleTxn.Uid, *tokenBundle.ID, *newSaleTxn.TransactionId)
	if err != nil {
		return nil, err
	}
	if updatedBalance.ID == nil {
		return nil, &core.ErrorResp{
			Message: "ERROR: updated user balance is invalid",
		}
	}

	return completedNewSaleTxn, nil
}

func (service *TransactionSvcImpl) ChargeTransaction(c context.Context, txn *model.Transaction, tokenService TokenService) (*model.Transaction, error) {
	// get the token amount for the bundle purchased
	bundle, err := tokenService.GetBundle(c, uint64(txn.TokenBundleId))
	if err != nil {
		return nil, err
	}

	if bundle.TokenAmount == nil {
		return nil, &core.ErrorResp{
			Message: "ERROR: token bundle is null",
		}
	}

	// updating billed initial price for transaction with dollar amount from bundle purchased
	txn.InitialPrice = strconv.FormatFloat(*bundle.DollarAmount, 'f', -1, 64)

	// build request to CCBill CBPT-API
	baseURL := os.Getenv("CCBILL_BILLING_API_URL")

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		log.Fatalf("Error parsing URL: %v", err)
	}

	// get CCBill datalink username
	datalinkUsername := os.Getenv("DATALINK_USER")

	// get CCBill datalink passowrd
	datalinkPasswordKey := os.Getenv("DATALINK_PASSWORD_KEY")
	datalinkPassword, err := core.GetSecret(datalinkPasswordKey)
	if err != nil {
		return nil, err
	}

	// get CCBill account numbers
	clientAccNum := os.Getenv("CCBILL_ACCNUM")
	clientSubAccnum := os.Getenv("CCBILL_SUBACCNUM")

	// Add query parameters
	params := url.Values{}
	params.Add("clientAccnum", clientAccNum)
	params.Add("username", datalinkUsername)
	params.Add("password", datalinkPassword)
	params.Add("action", "chargeByPreviousTransactionId")
	params.Add("newClientAccnum", clientAccNum)
	params.Add("newClientSubacc", clientSubAccnum)
	params.Add("sharedAuthentication", "1")
	params.Add("initialPrice", txn.InitialPrice)
	params.Add("initialPeriod", txn.InitialPeriod)
	params.Add("recurringPrice", txn.RecurringPrice)
	params.Add("recurringPeriod", txn.RecurringPeriod)
	params.Add("rebills", txn.Rebills)
	params.Add("subscriptionId", txn.SubscriptionId)
	params.Add("currencyCode", txn.CurrencyCode)

	// check for dev to add client sub acc
	env := os.Getenv("ENV")
	if env == "dev" {
		params.Add("clientSubacc", clientSubAccnum)
	}
	parsedURL.RawQuery = params.Encode()

	fmt.Println(parsedURL.String())

	requestBody := []byte(`{}`)
	req, err := http.NewRequest("GET", parsedURL.String(), bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// Set the content type to JSON
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	fmt.Println("Response: ", string(body))

	// parse CSV response
	reader := csv.NewReader(strings.NewReader(string(body)))
	records, err := reader.ReadAll()
	if err != nil {
		log.Printf("Error reading CSV: %v", err)
		return nil, nil
	}

	if len(records) > 1 {
		// checking for transaction approval
		if records[1][0] == "1" {
			txn.SubscriptionId = records[1][1]
			txn.TranDatetime = time.Now().Format("2006-01-02 15:04:05")

			// buy the tokens and update user token balance
			fmt.Println("Transaction ID: ", txn.TransactionId)
			updatedBalance, err := tokenService.BuyTokens(c, txn.Uid, uint64(txn.TokenBundleId), txn.TransactionId)
			if err != nil {
				return nil, err
			}
			if updatedBalance.ID == nil {
				return nil, &core.ErrorResp{
					Message: "ERROR: updated user balance is invalid",
				}
			}
			completedTxn, err := service.transactionRepo.ChargeTransaction(c, txn)
			if err != nil {
				return nil, err
			}

			completedTxn.TokenAmount = *bundle.TokenAmount
			return completedTxn, nil
		} else {
			return nil, &core.ErrorResp{
				Message: "An error occurred processing your transaction at this time. Please contact support@ccbill.com for assitance if this error persists.",
			}
		}
	} else {
		return nil, &core.ErrorResp{
			Message: "An error occurred processing your transaction at this time. Please contact support@ccbill.com for assitance if this error persists.",
		}
	}
}

func (service *TransactionSvcImpl) GetCharge(c context.Context) {
	service.transactionRepo.GetCharge(c)
}
