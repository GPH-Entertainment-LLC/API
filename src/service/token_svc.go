package service

import (
	"context"
	"fmt"
	"xo-packs/core"
	"xo-packs/model"
	"xo-packs/repository"
)

type TokenService interface {
	AddBundle(context.Context, *float64, *float64, string) (*model.TokenBundle, error)
	ActiveTokenRate(context.Context) (*model.TokenCurrencyRate, error)
	GetBundle(context.Context, uint64) (*model.TokenBundle, error)
	BuyTokens(context.Context, string, uint64, string) (*model.TokenBalance, error)
	GetBalance(context.Context, string) (*model.TokenBalance, error)
	GetCurrentBundles(context.Context) ([]*model.TokenBundle, error)
	GetBundlesByPrice(context.Context, *float64) ([]*model.TokenBundle, error)
	GetBundleByPrice(context.Context, *float64) (*model.TokenBundle, error)
	GetBundlesByPriceRange(context.Context, *float64, *float64) ([]*model.TokenBundle, error)
	GetInactiveBundles(context.Context) ([]*model.TokenBundle, error)
	UpdateBalance(context.Context, string, map[string]interface{}) (*model.TokenBalance, error)
	DeleteBundle(context.Context, uint64) (*model.TokenBundle, error)
	ClearUserTokenCache(context.Context, string) error
}

type TokenSvcImpl struct {
	tokenRepo repository.TokenRepository
}

func NewTokenService(tokenRepository repository.TokenRepository) TokenService {
	tokenService := TokenSvcImpl{tokenRepo: tokenRepository}
	return &tokenService
}

func (tokenService *TokenSvcImpl) AddBundle(c context.Context, dollarAmt *float64, tokenAmt *float64, bundleImageUrl string) (*model.TokenBundle, error) {
	return tokenService.tokenRepo.AddBundle(c, dollarAmt, tokenAmt, bundleImageUrl)
}

func (tokenService *TokenSvcImpl) ActiveTokenRate(c context.Context) (*model.TokenCurrencyRate, error) {
	return tokenService.tokenRepo.ActiveTokenRate(c)
}

func (tokenService *TokenSvcImpl) GetBundle(c context.Context, id uint64) (*model.TokenBundle, error) {
	return tokenService.tokenRepo.GetBundle(c, id)
}

func (tokenService *TokenSvcImpl) BuyTokens(c context.Context, uid string, tokenBundleId uint64, transactionId string) (*model.TokenBalance, error) {
	fmt.Println("INSIDE SERVICE: ", uid)
	tokenOrder, err := tokenService.tokenRepo.BuyTokens(c, uid, tokenBundleId, transactionId)
	if err != nil {
		return nil, err
	}

	if tokenOrder == nil {
		return nil, &core.ErrorResp{
			Message: "an error occurred processing the order",
		}
	}

	// invalidating user token cache
	if err = tokenService.ClearUserTokenCache(c, uid); err != nil {
		return nil, err
	}

	// getting new user token balance
	tokenBalance, err := tokenService.tokenRepo.GetBalance(c, uid)
	if err != nil {
		return nil, err
	}
	return tokenBalance, err
}

func (tokenService *TokenSvcImpl) GetBalance(c context.Context, uid string) (*model.TokenBalance, error) {
	return tokenService.tokenRepo.GetBalance(c, uid)
}

func (tokenService *TokenSvcImpl) GetCurrentBundles(c context.Context) ([]*model.TokenBundle, error) {
	return tokenService.tokenRepo.GetCurrentBundles(c)
}

func (tokenService *TokenSvcImpl) GetBundlesByPrice(c context.Context, dollarAmt *float64) ([]*model.TokenBundle, error) {
	return tokenService.tokenRepo.GetBundlesByPrice(c, dollarAmt)
}

func (tokenService *TokenSvcImpl) GetBundleByPrice(c context.Context, dollarAmt *float64) (*model.TokenBundle, error) {
	return tokenService.tokenRepo.GetBundleByPrice(c, dollarAmt)
}

func (tokenService *TokenSvcImpl) GetBundlesByPriceRange(c context.Context, lowerDollarAmt *float64, upperDollarAmt *float64) ([]*model.TokenBundle, error) {
	return tokenService.tokenRepo.GetBundlesByPriceRange(c, lowerDollarAmt, upperDollarAmt)
}

func (tokenService *TokenSvcImpl) GetInactiveBundles(c context.Context) ([]*model.TokenBundle, error) {
	return tokenService.tokenRepo.GetInactiveBundles(c)
}

func (tokenService *TokenSvcImpl) UpdateBalance(c context.Context, uid string, patchMap map[string]interface{}) (*model.TokenBalance, error) {
	balance, err := tokenService.tokenRepo.UpdateBalance(c, uid, patchMap)
	if err != nil {
		return nil, err
	}

	// invalidating user token cache
	if err = tokenService.ClearUserTokenCache(c, uid); err != nil {
		return nil, err
	}

	return balance, err
}

func (tokenService *TokenSvcImpl) DeleteBundle(c context.Context, id uint64) (*model.TokenBundle, error) {
	return tokenService.tokenRepo.DeleteBundle(c, id)
}

func (tokenService *TokenSvcImpl) ClearUserTokenCache(c context.Context, uid string) error {
	return tokenService.tokenRepo.ClearUserTokenCache(c, uid)
}
