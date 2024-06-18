package service

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"
	"xo-packs/core"
	"xo-packs/model"
	"xo-packs/repository"
)

type PackService interface {
	CreatePackConfig(context.Context, *model.PackConfig, VendorService) (*model.PackConfig, error)
	AddPackItemConfigs(context.Context, []*model.PackItemConfig) error
	GeneratePacks(context.Context, uint64, string, VendorService) error
	BuyPacks(context.Context, string, uint64, float64, TokenService) (*model.PackBoughtResp, error)
	OpenPack(context.Context, uint64, string, ItemService) (*model.Pack, error)
	GetPack(context.Context, uint64) (*model.Pack, error)
	AddPackCategories(context.Context, []*model.PackCategory) error
	GetUserPackAmount(context.Context, string) (*uint64, error)
	GetPackConfig(context.Context, uint64) (*model.PackConfig, error)
	GetPackItems(context.Context, uint64, ItemService) ([]*model.PackItemConfigExpanded, error)
	GetPackItemsPreview(context.Context, uint64, ItemService) ([]*model.PackItemConfigExpanded, error)
	GetPackItemsPreview2(context.Context, uint64) ([]*model.PackItemPreview, error)
	GetActivePackItems(context.Context, []uint64) ([]string, []string, error)
	GetPacksContainingItems(context.Context, []uint64) ([]string, []string, error)
	PatchPackConfig(context.Context, uint64, map[string]interface{}, string) (*model.PackConfig, error)
	RemoveUserPacks(context.Context, []uint64) error
	ActivatePacks(context.Context, []uint64, string, VendorService) error
	DeactivatePacks(context.Context, []uint64, string, VendorService) error
	DeletePackConfigs(context.Context, []uint64, string, VendorService) error
	ClearPackCategoryCache(context.Context) error
	ClearVendorPackCache(context.Context, string) error
	ClearPackShopCache(context.Context) error
	ClearPackConfigCache(context.Context, []uint64, string) error
	GeneratePackItemOdds(context.Context, string, []model.PackItemConfig, int, ItemService) ([]int, error)
}

type PackSvcImpl struct {
	packRepo repository.PackRepository
}

func NewPackService(repo repository.PackRepository) PackService {
	return &PackSvcImpl{packRepo: repo}
}

// @service: create-pack-config
func (packService *PackSvcImpl) CreatePackConfig(c context.Context, packConfig *model.PackConfig, vendorService VendorService) (*model.PackConfig, error) {
	if packConfig.Qty == nil || packConfig.ItemQty == nil || packConfig.TokenAmount == nil {
		return nil, &core.ErrorResp{
			Message: "pack qty, item qty, and token amount need to be populated in the pack config",
		}
	}

	if (*packConfig.Qty * *packConfig.ItemQty) > 10000 {
		return nil, &core.ErrorResp{
			Message: "total pack items cannot exceed 10,000",
		}
	}

	// checking token amount is high enough
	if *packConfig.TokenAmount < 5 {
		return nil, &core.ErrorResp{
			Message: "pack must cost at least 5 tokens",
		}
	}

	if packConfig.VendorID == nil {
		return nil, &core.ErrorResp{
			Message: "vendor id cannot be null",
		}
	}
	vendorId := *packConfig.VendorID

	// converting release date time to UTC for scheduler job
	// if packConfig.ReleaseAt != nil {
	// 	parsedTime, err := time.Parse(time.RFC3339, *packConfig.ReleaseAt)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	// Convert the parsed time to UTC timezone
	// 	*packConfig.ReleaseAt =
	// }

	packConfig, err := packService.packRepo.CreatePackConfig(c, packConfig)
	if err != nil {
		return nil, err
	}

	// clearing vendor pack cache
	if err := packService.ClearVendorPackCache(c, vendorId); err != nil {
		return nil, err
	}

	// clear vendor cache
	if err := vendorService.ClearVendorCache(c, vendorId); err != nil {
		return nil, err
	}

	return packConfig, nil
}

func (packService *PackSvcImpl) AddPackItemConfigs(c context.Context, packItemConfigs []*model.PackItemConfig) error {
	if len(packItemConfigs) > 0 {
		return packService.packRepo.AddPackItemConfigs(c, packItemConfigs)
	}
	return nil
}

func (packService *PackSvcImpl) GeneratePacks(c context.Context, packConfigId uint64, vendorId string, vendorService VendorService) error {
	// 1. get the pack config
	packConfig, err := packService.packRepo.GetPackConfig(c, packConfigId)
	if err != nil {
		return err
	}
	if packConfig == nil {
		return &core.SvcError{Message: "This pack has been discontinued"}
	} else if packConfig.DeletedAt != nil {
		return &core.SvcError{Message: "This pack has been discontinued"}
	}

	if packConfig.CurrentStock == nil {
		return &core.SvcError{
			Message: "Data quality error; no current stock associated with pack",
		}
	} else if *packConfig.CurrentStock > 0 {
		return &core.SvcError{
			Message: fmt.Sprintf("There are currently %v packs available in the market", *packConfig.CurrentStock),
		}
	}

	if packConfig.VendorID == nil {
		return &core.ErrorResp{Message: "this pack does not exist"}
	}
	if *packConfig.VendorID != vendorId {
		return &core.ErrorResp{Message: "vendor does not have access to this pack"}
	}

	// 2. get the pack item configs
	packItemConfigs, err := packService.packRepo.GetPackItemConfigs(c, packConfigId)
	if err != nil {
		return err
	}
	if len(packItemConfigs) == 0 {
		return &core.SvcError{Message: "This pack has no items associated with it"}
	}

	// 3. generate list of item ids for a new pack based on the configs
	packItemIdBatch, err := GeneratePackItemIds(c, packItemConfigs, packConfig)
	if err != nil {
		return err
	}
	if len(packItemIdBatch) == 0 {
		return &core.SvcError{Message: "An error occurred generating the item ids list"}
	}

	// 4. build and upload pack fact records
	packs := make([]*model.PackFact, *packConfig.Qty)
	for i := 0; i < *packConfig.Qty; i++ {
		active := true
		pack := model.PackFact{PackConfigID: &packConfigId, Active: &active}
		packs[i] = &pack
	}
	packIds, err := packService.packRepo.UploadPacks(c, packs, packConfigId)
	if err != nil {
		return &core.SvcError{Message: "An error occurred uploading packs to document store"}
	}

	if len(packIds) != len(packItemIdBatch) {
		return &core.SvcError{Message: fmt.Sprintf("Critical error: amount of packs uploaded and item id batches are not equal. Pack Config ID: %v", *packConfig.ID)}
	}

	// 5. build and upload pack item fact records
	items := []*model.PackItemFact{}
	for i, itemIds := range packItemIdBatch {
		for _, v := range itemIds {
			id := v
			item := model.PackItemFact{PackID: &packIds[i], ItemID: &id}
			items = append(items, &item)
		}
	}
	err = packService.packRepo.UploadPackItems(c, items)
	if err != nil {
		return err
	}

	// clear vendor pack cache
	err = packService.ClearVendorPackCache(c, vendorId)
	if err != nil {
		return err
	}

	// clear vendor cache
	err = vendorService.ClearVendorCache(c, vendorId)
	if err != nil {
		return err
	}

	return nil
}

func (packService *PackSvcImpl) AddPackCategories(c context.Context, categories []*model.PackCategory) error {
	if err := packService.ClearPackCategoryCache(c); err != nil {
		return err
	}

	return packService.packRepo.AddPackCategories(c, categories)
}

func (packService *PackSvcImpl) GetUserPackAmount(c context.Context, uid string) (*uint64, error) {
	return packService.packRepo.GetUserPackAmount(c, uid)
}

// function that associates a new pack fact with a user (updates owner ID of next available pack)
func (packService *PackSvcImpl) BuyPacks(c context.Context, uid string, packConfigId uint64, amount float64, tokenService TokenService) (*model.PackBoughtResp, error) {
	if amount <= 0 {
		err := &core.ErrorResp{Message: "cannot purchase 0 packs"}
		return nil, err
	}

	packConfig, err := packService.GetPackConfig(c, packConfigId)
	if err != nil {
		return nil, err
	}
	if packConfig.VendorID == nil || packConfig.ID == nil {
		err := &core.ErrorResp{Message: "critical error; pack config does not exist or vendor_id not present"}
		return nil, err
	}
	if packConfig.TokenAmount == nil {
		err := &core.ErrorResp{Message: "critical error; pack config has null token amount"}
		return nil, err
	}

	activeTokenRate, err := tokenService.ActiveTokenRate(c)
	if err != nil {
		return nil, err
	}

	resp, err := packService.packRepo.BuyPacks(c, uid, packConfig, amount, activeTokenRate)
	if err != nil {
		return nil, err
	}

	// invalidate user token balance cache
	if err := tokenService.ClearUserTokenCache(c, uid); err != nil {
		return nil, err
	}

	// remove vendor pack cache
	if err := packService.ClearVendorPackCache(c, *packConfig.VendorID); err != nil {
		return nil, err
	}

	// remove user pack cache
	if err := packService.ClearUserPackCache(c, uid); err != nil {
		return nil, err
	}

	// remove pack config cache
	if err := packService.ClearPackConfigCache(c, []uint64{packConfigId}, *packConfig.VendorID); err != nil {
		return nil, err
	}

	return resp, nil
}

func (packService *PackSvcImpl) GetPack(c context.Context, id uint64) (*model.Pack, error) {
	return packService.packRepo.GetPack(c, id)
}

func (packService *PackSvcImpl) OpenPack(c context.Context, id uint64, uid string, itemService ItemService) (*model.Pack, error) {
	if err := packService.ClearPackCache(c, id); err != nil {
		return nil, err
	}
	pack, err := packService.packRepo.GetPack(c, id)
	if err != nil {
		return nil, err
	}
	if pack.OwnerId == nil {
		return nil, &core.ErrorResp{Message: "Error: pack has no owner associated"}
	}
	if pack.OpenedAt != nil {
		return nil, &core.ErrorResp{Message: "Error: pack has already been opened"}
	}
	if *pack.OwnerId != uid {
		return nil, &core.ErrorResp{Message: "Error: user does not own this pack"}
	}

	pack, err = packService.packRepo.OpenPack(c, id, uid)
	if err != nil {
		return nil, err
	}
	if len(pack.Items) <= 0 {
		return nil, &core.ErrorResp{Message: fmt.Sprintf("Critical server error: pack %v is empty", id)}
	}

	// build batch url sign object for content service
	urlBatch := map[int]map[string]*string{}
	for _, item := range pack.Items {
		if _, ok := urlBatch[int(*item.ID)]; !ok {
			urlBatch[int(*item.ID)] = map[string]*string{
				"contentMainUrl":  item.ContentMainUrl,
				"contentThumbUrl": item.ContentThumbUrl,
			}
		}
	}

	pemKey, keyPairID, err := itemService.GetItemSignCreds(c)
	if err != nil {
		return nil, err
	}
	signedUrlBatch, err := core.SignUrlBatch(urlBatch, pemKey, keyPairID)
	if err != nil {
		return nil, err
	}

	for _, item := range pack.Items {
		item.ContentMainUrl = signedUrlBatch[int(*item.ID)]["contentMainUrl"]
		item.ContentThumbUrl = signedUrlBatch[int(*item.ID)]["contentThumbUrl"]
	}

	itemIds := []uint64{}
	for _, v := range pack.Items {
		itemIds = append(itemIds, *v.ID)
	}
	err = itemService.AddUserItems(c, uid, itemIds)
	if err != nil {
		return nil, err
	}
	return pack, err
}

// function to randomly generate item ids to associate with a new pack instance that a customer purchases
// runtime = O(N*M) where N = pack config item qty, M = possible items in pack
func GeneratePackItemIds(c context.Context, packItemConfigs []*model.PackItemConfig, packConfig *model.PackConfig) ([][]uint64, error) {
	packItemIdBatch := [][]uint64{}
	itemIdPool := []uint64{}

	// get the pool ready
	for _, v := range packItemConfigs {
		itemIds := make([]uint64, *v.Qty)
		for i := range itemIds {
			itemIds[i] = *v.ItemID
		}
		itemIdPool = append(itemIdPool, itemIds...)
	}

	// iterate through pack qty to build a list of item ids for each pack
	for i := 0; i < *packConfig.Qty; i++ {
		packItemIds := []uint64{}
		for j := 0; j < *packConfig.ItemQty; j++ {
			if len(itemIdPool) != 0 {
				rng := rand.New(rand.NewSource(time.Now().UnixNano()))
				randNum := rng.Intn(len(itemIdPool))
				packItemIds = append(packItemIds, itemIdPool[randNum])
				itemIdPool = append(itemIdPool[:randNum], itemIdPool[randNum+1:]...)
			}
		}
		packItemIdBatch = append(packItemIdBatch, packItemIds)
	}
	return packItemIdBatch, nil
}

func (packService *PackSvcImpl) GetPackConfig(c context.Context, id uint64) (*model.PackConfig, error) {
	packConfig, err := packService.packRepo.GetPackConfig(c, id)
	if err != nil {
		return nil, err
	}
	if packConfig.VendorID == nil {
		return nil, &core.ErrorResp{
			Message: "pack does not exist",
		}

	}
	return packConfig, err
}

func (packService *PackSvcImpl) GetPackItems(c context.Context, id uint64, itemService ItemService) ([]*model.PackItemConfigExpanded, error) {
	packItems, err := packService.packRepo.GetPackItems(c, id)
	if err != nil {
		return nil, err
	}

	// build batch url sign object for content service
	urlBatch := map[int]map[string]*string{}
	for _, item := range packItems {
		if _, ok := urlBatch[int(*item.ID)]; !ok {
			urlBatch[int(*item.ID)] = map[string]*string{
				"contentMainUrl":  item.ContentMainUrl,
				"contentThumbUrl": item.ContentThumbUrl,
			}
		}
	}

	pemKey, keyPairID, err := itemService.GetItemSignCreds(c)
	if err != nil {
		return nil, err
	}
	signedUrlBatch, err := core.SignUrlBatch(urlBatch, pemKey, keyPairID)
	if err != nil {
		return nil, err
	}

	for _, item := range packItems {
		item.ContentMainUrl = signedUrlBatch[int(*item.ID)]["contentMainUrl"]
		item.ContentThumbUrl = signedUrlBatch[int(*item.ID)]["contentThumbUrl"]
	}

	return packItems, nil
}

func (packService *PackSvcImpl) GetPackItemsPreview2(c context.Context, packConfigId uint64) ([]*model.PackItemPreview, error) {
	return packService.packRepo.GetPackItemsPreview(c, packConfigId)
}

func (packService *PackSvcImpl) GetPackItemsPreview(c context.Context, id uint64, itemService ItemService) ([]*model.PackItemConfigExpanded, error) {
	packItems, err := packService.packRepo.GetPackItems(c, id)
	if err != nil {
		return nil, err
	}

	// build batch url sign object for content service
	urlBatch := map[int]map[string]*string{}
	for _, item := range packItems {
		if _, ok := urlBatch[int(*item.ID)]; !ok {
			urlBatch[int(*item.ID)] = map[string]*string{
				"contentMainUrl":  item.ContentMainUrl,
				"contentThumbUrl": item.ContentThumbUrl,
			}
		}
	}

	pemKey, keyPairID, err := itemService.GetItemSignCreds(c)
	if err != nil {
		return nil, err
	}
	signedUrlBatch, err := core.SignUrlBatch(urlBatch, pemKey, keyPairID)
	if err != nil {
		return nil, err
	}

	for _, item := range packItems {
		item.ContentMainUrl = signedUrlBatch[int(*item.ID)]["contentMainUrl"]
		item.ContentThumbUrl = signedUrlBatch[int(*item.ID)]["contentThumbUrl"]
	}

	return packItems, nil
}

func (packService *PackSvcImpl) GetActivePackItems(c context.Context, itemIds []uint64) ([]string, []string, error) {
	return packService.packRepo.GetActivePackItems(c, itemIds)
}

func (packService *PackSvcImpl) GetPacksContainingItems(c context.Context, itemIds []uint64) ([]string, []string, error) {
	return packService.packRepo.GetPacksContainingItems(c, itemIds)
}

func (packService *PackSvcImpl) PatchPackConfig(c context.Context, packConfigId uint64, packConfigPatchMap map[string]interface{}, vendorId string) (*model.PackConfig, error) {
	packConfigPatchMap = core.ConvertJSONMapToDBMap(packConfigPatchMap, model.PackConfig{})

	// // converting release date time to UTC for scheduler job
	// if packConfigPatchMap["releaseAt"] != nil {
	// 	parsedTime, err := time.Parse(time.RFC3339, packConfigPatchMap["releaseAt"].(string))
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	packConfigPatchMap["releaseAt"] = parsedTime.String()
	// }

	err := packService.packRepo.PatchPackConfig(c, packConfigId, packConfigPatchMap, vendorId)
	if err != nil {
		return nil, err
	}

	// clear pack config
	err = packService.ClearPackConfigCache(c, []uint64{packConfigId}, vendorId)
	if err != nil {
		return nil, err
	}

	packConfig, err := packService.packRepo.GetPackConfig(c, packConfigId)
	if err != nil {
		return nil, err
	}

	return packConfig, err
}

func (packService *PackSvcImpl) RemoveUserPacks(c context.Context, packIds []uint64) error {
	return nil
}

func (packService *PackSvcImpl) ActivatePacks(c context.Context, packConfigIds []uint64, vendorId string, vendorService VendorService) error {
	err := packService.packRepo.ActivatePacks(c, packConfigIds, vendorId)
	if err != nil {
		return err
	}

	// clearing vendor pack cache
	if err := packService.ClearVendorPackCache(c, vendorId); err != nil {
		return err
	}

	// clear vendor cache
	if err := vendorService.ClearVendorCache(c, vendorId); err != nil {
		return err
	}
	return nil
}

func (packService *PackSvcImpl) DeactivatePacks(c context.Context, packConfigIds []uint64, vendorId string, vendorService VendorService) error {
	err := packService.packRepo.DeactivatePacks(c, packConfigIds, vendorId)
	if err != nil {
		return err
	}

	// clearing vendor pack cache
	if err := packService.ClearVendorPackCache(c, vendorId); err != nil {
		return err
	}

	// clear vendor cache
	if err := vendorService.ClearVendorCache(c, vendorId); err != nil {
		return err
	}
	return nil
}

func (packService *PackSvcImpl) DeletePackConfigs(c context.Context, packConfigIds []uint64, vendorId string, vendorService VendorService) error {
	err := packService.packRepo.DeletePackConfigs(c, packConfigIds, vendorId)
	if err != nil {
		return err
	}

	// clearing pack configs
	if err := packService.ClearPackConfigCache(c, packConfigIds, vendorId); err != nil {
		return err
	}

	// clearing vendor pack cache
	if err := packService.ClearVendorPackCache(c, vendorId); err != nil {
		return err
	}

	// clear vendor cache
	if err := vendorService.ClearVendorCache(c, vendorId); err != nil {
		return err
	}
	return nil
}

func (packService *PackSvcImpl) ClearPackConfigCache(c context.Context, packConfigIds []uint64, vendorId string) error {
	return packService.packRepo.ClearPackConfigCache(c, packConfigIds, vendorId)
}

func (packService *PackSvcImpl) ClearPackCategoryCache(c context.Context) error {
	return packService.packRepo.ClearPackCategoryCache(c)
}

func (packService *PackSvcImpl) ClearUserPackCache(c context.Context, uid string) error {
	return packService.packRepo.ClearUserPackCache(c, uid)
}

func (packService *PackSvcImpl) ClearVendorPackCache(c context.Context, vendorId string) error {
	return packService.packRepo.ClearVendorPackCache(c, vendorId)
}

func (packService *PackSvcImpl) ClearPackCache(c context.Context, packId uint64) error {
	return packService.packRepo.ClearPackCache(c, packId)
}

func (packService *PackSvcImpl) ClearPackShopCache(c context.Context) error {
	return packService.packRepo.ClearPackShopCache(c)
}

func (packService *PackSvcImpl) GeneratePackItemOdds(c context.Context, vendorId string, itemConfigs []model.PackItemConfig, totalItems int, itemService ItemService) ([]int, error) {
	// get the list of items from DB
	itemIds := []uint64{}
	for _, itemConfig := range itemConfigs {
		if itemConfig.ItemID != nil {
			itemIds = append(itemIds, *itemConfig.ItemID)
		} else {
			return nil, &core.ErrorResp{
				Message: fmt.Sprintf("Critical error: an item config does not have an item id attached"),
			}
		}
	}
	items, err := itemService.GetItems(c, itemIds, vendorId)
	if err != nil {
		return nil, err
	}

	// sort the items in asc order on rarity ids
	sort.Slice(items, func(i, j int) bool {
		return *items[i].RarityId < *items[j].RarityId
	})

	// generate odds for items
	itemOddsMap, err := packService.packRepo.GeneratePackItemOdds(c, items, totalItems)
	if err != nil {
		return nil, err
	}

	oddsSlice := []int{}

	// building odds distribution list with the original input item order
	for _, itemId := range itemIds {
		qty, ok := itemOddsMap[itemId]
		if !ok {
			return nil, &core.ErrorResp{
				Message: fmt.Sprintf("Critical error: item odds map has no item id amount entry for item id: %v", itemId),
			}
		}
		oddsSlice = append(oddsSlice, qty)
	}

	return oddsSlice, nil
}
