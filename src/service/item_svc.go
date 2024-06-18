package service

import (
	"context"
	"fmt"
	"xo-packs/core"
	"xo-packs/db"
	"xo-packs/model"
	"xo-packs/repository"
)

type ItemService interface {
	CreateItem(context.Context, *model.Item) (*model.Item, error)
	GetItem(context.Context, uint64) (*model.Item, error)
	GetItems(context.Context, []uint64, string) ([]model.Item, error)
	AddItemCategories(context.Context, []*model.ItemCategory) error
	AddUserItems(context.Context, string, []uint64) error
	PatchItem(context.Context, uint64, map[string]interface{}, string) (*model.Item, error)
	DeleteUserItems(context.Context, []uint64, string) error
	DeleteItems(context.Context, []uint64, string, PackService) error
	ClearVendorItemCache(context.Context, string) error
	ClearItemCategoryCache(context.Context) error
	ClearUserItemCache(context.Context, string) error
	ClearItemCache(context.Context, []uint64) error
	GetItemSignCreds(context.Context) (string, string, error)
}

type ItemSvcImpl struct {
	itemRepo repository.ItemRepository
}

func NewItemService(repo repository.ItemRepository) ItemService {
	return &ItemSvcImpl{itemRepo: repo}
}

type ItemError struct {
	message string
}

func (e *ItemError) Error() string {
	return e.message
}

func (itemService *ItemSvcImpl) CreateItem(c context.Context, item *model.Item) (*model.Item, error) {
	item, err := itemService.itemRepo.CreateItem(c, item)
	if err != nil {
		return nil, err
	}

	// remove cached vendor item list
	if err := itemService.ClearVendorItemCache(c, *item.VendorId); err != nil {
		return nil, err
	}

	return item, err
}

func (itemService *ItemSvcImpl) GetItem(c context.Context, itemId uint64) (*model.Item, error) {
	item, err := itemService.itemRepo.GetItem(c, itemId)
	if err != nil {
		return nil, err
	}
	if item.VendorId == nil {
		return nil, &core.ErrorResp{
			Message: "item does not exist",
		}
	}
	return item, err
}

func (itemService *ItemSvcImpl) GetItems(c context.Context, itemIds []uint64, vendorId string) ([]model.Item, error) {
	return itemService.itemRepo.GetItems(c, itemIds, vendorId)
}

func (itemService *ItemSvcImpl) AddItemCategories(c context.Context, categories []*model.ItemCategory) error {
	err := itemService.itemRepo.AddItemCategories(c, categories)
	if err != nil {
		return err
	}

	// removing cached item categories
	if err := itemService.ClearItemCategoryCache(c); err != nil {
		return err
	}
	return err
}

func (itemService *ItemSvcImpl) AddUserItems(c context.Context, uid string, itemIds []uint64) error {
	err := itemService.itemRepo.AddUserItems(c, uid, itemIds)
	if err != nil {
		return err
	}

	// clearing user item cache
	return itemService.ClearUserItemCache(c, uid)
}

func (itemService *ItemSvcImpl) PatchItem(c context.Context, itemId uint64, itemPatchMap map[string]interface{}, authorizedUid string) (*model.Item, error) {
	dbPatchMap := core.ConvertJSONMapToDBMap(itemPatchMap, model.Item{})
	updatedItem, err := itemService.itemRepo.PatchItem(c, itemId, dbPatchMap, authorizedUid)
	if err != nil {
		return nil, err
	}

	// clear vendor item cache
	if err = itemService.ClearVendorItemCache(c, authorizedUid); err != nil {
		return nil, err
	}

	return updatedItem, err
}

func (itemService *ItemSvcImpl) DeleteUserItems(c context.Context, userItemIds []uint64, uid string) error {
	err := itemService.itemRepo.DeleteUserItems(c, userItemIds, uid)
	if err != nil {
		return err
	}

	// removing user item cache
	if err := itemService.ClearUserItemCache(c, uid); err != nil {
		return err
	}
	return err
}

func (itemService *ItemSvcImpl) DeleteItems(c context.Context, itemIds []uint64, vendorId string, packService PackService) error {
	packNames, packTitles, err := packService.GetPacksContainingItems(c, itemIds)
	if err != nil {
		return err
	}

	if len(packNames) > 0 {
		return &core.ErrorResp{Message: fmt.Sprintf("Item '%v' belongs to pack '%v' and must be deleted before proceeding", packNames[0], packTitles[0])}
	}

	err = itemService.itemRepo.DeleteItems(c, itemIds, vendorId)
	if err != nil {
		return err
	}

	if err = itemService.ClearItemCache(c, itemIds); err != nil {
		return err
	}

	// removing cached vendor item pages
	if err = itemService.ClearVendorItemCache(c, vendorId); err != nil {
		return err
	}
	return err
}

func (itemService *ItemSvcImpl) ClearItemCategoryCache(c context.Context) error {
	return itemService.itemRepo.ClearItemCategoryCache(c)
}

func (itemService *ItemSvcImpl) ClearUserItemCache(c context.Context, uid string) error {
	return itemService.itemRepo.ClearUserItemCache(c, uid)
}

func (itemService *ItemSvcImpl) ClearVendorItemCache(c context.Context, vendorId string) error {
	return itemService.itemRepo.ClearVendorItemCache(c, vendorId)
}

func (itemService *ItemSvcImpl) ClearItemCache(c context.Context, itemIds []uint64) error {
	// removing cached items
	itemKeys := []string{}
	for _, itemId := range itemIds {
		itemKeys = append(itemKeys, fmt.Sprintf("%v%v", db.KEY_ITEM, itemId))
	}

	return itemService.itemRepo.ClearItemCache(c, itemKeys)
}

func (itemService *ItemSvcImpl) GetItemSignCreds(c context.Context) (string, string, error) {
	return itemService.itemRepo.GetItemSignCreds(c)
}
