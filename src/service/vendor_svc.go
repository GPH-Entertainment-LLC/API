package service

import (
	"context"
	"math"
	"os"
	"regexp"
	"slices"
	"strconv"
	"xo-packs/core"
	"xo-packs/model"
	"xo-packs/repository"
)

type VendorService interface {
	GetVendorSortList(context.Context, string) ([]string, error)
	GetActiveVendor(context.Context, string) (*model.Vendor, error)
	GetVendor(context.Context, string) (*model.Vendor, error) // NOTE: should replace get vendor once confirmed and transitioned to GetActiveVendor
	GetVendorAmount(context.Context) (*uint64, error)
	GetVendorsPage(context.Context, uint64, string, string, string, string, string, CategoryService) (*model.VendorPage, error)
	GetVendorShopPackPage(context.Context, uint64, string, string, string, string, string, CategoryService) (*model.VendorPackPage, error)                      // shop user view
	GetVendorProfilePackPage(context.Context, string, uint64, string, string, string, string, CategoryService) (*model.VendorPackPage, error)                   // vendor profile user view
	GetVendorPackListPage(context.Context, string, uint64, string, string, string, string, string, CategoryService) (*model.VendorPackPage, error)              // vendor pack list vendor view
	GetVendorItemListPage(context.Context, string, uint64, string, string, string, string, string, CategoryService, ItemService) (*model.VendorItemPage, error) // vendor item list vendor view
	GetVendorCategories(context.Context, string) ([]*model.VendorCategoryExpanded, error)
	AddVendorCategories(context.Context, []*model.VendorCategory) error
	ApproveVendor(context.Context, string) (*model.Vendor, error)
	RemoveVendor(context.Context, string) error
	PatchVendor(context.Context, string, map[string]interface{}) (*model.Vendor, error)
	ClearVendorCache(context.Context, string) error
}

type VendorSvcImpl struct {
	vendorRepo repository.VendorRepository
}

func NewVendorService(vendorRepo repository.VendorRepository) VendorService {
	return &VendorSvcImpl{vendorRepo: vendorRepo}
}

func (vendorService *VendorSvcImpl) GetVendorSortList(c context.Context, urlPath string) ([]string, error) {
	return vendorService.vendorRepo.GetVendorSortList(c, urlPath)
}

func (vendorService *VendorSvcImpl) GetVendorAmount(c context.Context) (*uint64, error) {
	return vendorService.vendorRepo.GetVendorAmount(c)
}

func (vendorService *VendorSvcImpl) GetVendorsPage(
	c context.Context, pageNumber uint64, sortBy string, sortDir string, filterOn string, searchStr string, urlPath string, categoryService CategoryService) (*model.VendorPage, error) {

	if filterOn != "" {
		categories, err := categoryService.GetCategoryLiterals(c)
		if err != nil {
			return nil, err
		}
		if !slices.Contains(categories, filterOn) {
			return nil, &core.ErrorResp{Message: "Filter on category not valid"}
		}
	}
	if sortBy != "" {
		sortMapping, err := vendorService.vendorRepo.GetVendorSortMappings(c)
		if err != nil {
			return nil, err
		}
		_, ok := sortMapping[sortBy]
		if !ok {
			return nil, &core.ErrorResp{Message: "Sort by parameter not a valid mapping param"}
		}
		if sortDir != "asc" && sortDir != "desc" {
			return nil, &core.ErrorResp{Message: "sortDir param must be 'asc' or 'desc'"}
		}
		sortBy = sortMapping[sortBy] + " " + sortDir
	}

	adjustedPageNum := pageNumber - 1
	vendorsPage, vendorAmount, err := vendorService.vendorRepo.GetVendorsPage(c, adjustedPageNum, sortBy, filterOn, searchStr, urlPath)
	if err != nil {
		return nil, err
	}
	pageSizeStr := os.Getenv("VENDOR_PAGE_SIZE")
	pageSize, err := strconv.ParseUint(pageSizeStr, 10, 64)
	if err != nil {
		return nil, err
	}
	thisPageSize := uint64(len(vendorsPage))
	nextPageNum := uint64(0)
	nextPage := &nextPageNum
	if thisPageSize < pageSize {
		nextPage = nil
	} else {
		nextPageNum = (pageNumber + uint64(1))
	}
	result := model.VendorPage{VendorAmount: vendorAmount, PageSize: &thisPageSize, NextPage: nextPage, Page: vendorsPage}
	return &result, nil
}

func (vendorService *VendorSvcImpl) GetVendorShopPackPage(
	c context.Context, pageNumber uint64, sortBy string, sortDir string, filterOn string, searchStr string, urlPath string, categoryService CategoryService) (*model.VendorPackPage, error) {

	if filterOn != "" {
		categories, err := categoryService.GetCategoryLiterals(c)
		if err != nil {
			return nil, err
		}
		if !slices.Contains(categories, filterOn) {
			return nil, &core.ErrorResp{Message: "Filter on category not valid"}
		}
	}
	if sortBy != "" {
		sortMapping, err := vendorService.vendorRepo.GetVendorShopPackSortMappings(c)
		if err != nil {
			return nil, err
		}
		_, ok := sortMapping[sortBy]
		if !ok {
			return nil, &core.ErrorResp{Message: "Sort by parameter not a valid mapping param"}
		}
		if sortDir != "asc" && sortDir != "desc" {
			return nil, &core.ErrorResp{Message: "sortDir param must be 'asc' or 'desc'"}
		}
		sortBy = sortMapping[sortBy] + " " + sortDir
	}

	adjustedPageNum := pageNumber - 1
	vendorShopPackPage, vendorPackAmount, err := vendorService.vendorRepo.GetVendorShopPackPage(c, adjustedPageNum, sortBy, filterOn, searchStr, urlPath)
	if err != nil {
		return nil, err
	}
	pageSizeStr := os.Getenv("VENDOR_PACK_SHOP_PAGE_SIZE")
	pageSize, err := strconv.ParseUint(pageSizeStr, 10, 64)
	if err != nil {
		return nil, err
	}
	thisPageSize := uint64(len(vendorShopPackPage))
	nextPageNum := uint64(0)
	nextPage := &nextPageNum
	if thisPageSize < pageSize {
		nextPage = nil
	} else {
		nextPageNum = (pageNumber + uint64(1))
	}
	result := model.VendorPackPage{VendorPackAmount: vendorPackAmount, PageSize: &thisPageSize, NextPage: nextPage, Page: vendorShopPackPage}
	return &result, nil
}

func (vendorService *VendorSvcImpl) GetVendorProfilePackPage(
	c context.Context, uid string, pageNumber uint64, sortBy string, sortDir string, filterOn string, urlPath string, categoryService CategoryService) (*model.VendorPackPage, error) {

	if filterOn != "" {
		categories, err := categoryService.GetCategoryLiterals(c)
		if err != nil {
			return nil, err
		}
		if !slices.Contains(categories, filterOn) {
			return nil, &core.ErrorResp{Message: "Filter on category not valid"}
		}
	}
	if sortBy != "" {
		sortMapping, err := vendorService.vendorRepo.GetVendorPackSortMappings(c)
		if err != nil {
			return nil, err
		}
		_, ok := sortMapping[sortBy]
		if !ok {
			return nil, &core.ErrorResp{Message: "Sort by parameter not a valid mapping param"}
		}
		if sortDir != "asc" && sortDir != "desc" {
			return nil, &core.ErrorResp{Message: "sortDir param must be 'asc' or 'desc'"}
		}
		sortBy = sortMapping[sortBy] + " " + sortDir
	}

	adjustedPageNum := pageNumber - 1
	vendorProfilePackPage, vendorPackAmount, err := vendorService.vendorRepo.GetVendorProfilePackPage(c, uid, adjustedPageNum, sortBy, filterOn, urlPath)
	if err != nil {
		return nil, err
	}
	pageSizeStr := os.Getenv("VENDOR_PROFILE_PACK_PAGE_SIZE")
	pageSize, err := strconv.ParseUint(pageSizeStr, 10, 64)
	if err != nil {
		return nil, err
	}
	thisPageSize := uint64(len(vendorProfilePackPage))
	nextPageNum := uint64(0)
	nextPage := &nextPageNum
	if thisPageSize < pageSize {
		nextPage = nil
	} else {
		nextPageNum = (pageNumber + uint64(1))
	}
	result := model.VendorPackPage{VendorPackAmount: vendorPackAmount, PageSize: &thisPageSize, NextPage: nextPage, Page: vendorProfilePackPage}
	return &result, nil
}

func (vendorService *VendorSvcImpl) GetVendorPackListPage(
	c context.Context, uid string, pageNumber uint64, sortBy string, sortDir string, filterOn string, searchStr string, urlPath string, categoryService CategoryService) (*model.VendorPackPage, error) {

	if filterOn != "" {
		categories, err := categoryService.GetCategoryLiterals(c)
		if err != nil {
			return nil, err
		}
		if !slices.Contains(categories, filterOn) {
			return nil, &core.ErrorResp{Message: "Filter on category not valid"}
		}
	}
	if sortBy != "" {
		sortMapping, err := vendorService.vendorRepo.GetVendorPackSortMappings(c)
		if err != nil {
			return nil, err
		}
		_, ok := sortMapping[sortBy]
		if !ok {
			return nil, &core.ErrorResp{Message: "Sort by parameter not a valid mapping param"}
		}
		if sortDir != "asc" && sortDir != "desc" {
			return nil, &core.ErrorResp{Message: "sortDir param must be 'asc' or 'desc'"}
		}
		sortBy = sortMapping[sortBy] + " " + sortDir
	}

	adjustedPageNum := pageNumber - 1
	vendorPackListPage, vendorPackAmount, err := vendorService.vendorRepo.GetVendorPackListPage(c, uid, adjustedPageNum, sortBy, filterOn, searchStr, urlPath)
	if err != nil {
		return nil, err
	}
	pageSizeStr := os.Getenv("VENDOR_PACKS_LIST_PAGE_SIZE")
	pageSize, err := strconv.ParseUint(pageSizeStr, 10, 64)
	if err != nil {
		return nil, err
	}
	thisPageSize := uint64(len(vendorPackListPage))
	nextPageNum := uint64(0)
	nextPage := &nextPageNum
	if thisPageSize < pageSize {
		nextPage = nil
	} else {
		nextPageNum = (pageNumber + uint64(1))
	}
	result := model.VendorPackPage{VendorPackAmount: vendorPackAmount, PageSize: &thisPageSize, NextPage: nextPage, Page: vendorPackListPage}
	return &result, nil
}

func (vendorService *VendorSvcImpl) GetVendorItemListPage(
	c context.Context, uid string, pageNumber uint64, sortBy string, sortDir string, filterOn string, searchStr string, urlPath string, categoryService CategoryService, itemService ItemService) (*model.VendorItemPage, error) {

	if filterOn != "" {
		categories, err := categoryService.GetCategoryLiterals(c)
		if err != nil {
			return nil, err
		}
		if !slices.Contains(categories, filterOn) {
			return nil, &core.ErrorResp{Message: "Filter on category not valid"}
		}
	}
	if sortBy != "" {
		sortMapping, err := vendorService.vendorRepo.GetVendorItemSortMappings(c)
		if err != nil {
			return nil, err
		}
		_, ok := sortMapping[sortBy]
		if !ok {
			return nil, &core.ErrorResp{Message: "Sort by parameter not a valid mapping param"}
		}
		if sortDir != "asc" && sortDir != "desc" {
			return nil, &core.ErrorResp{Message: "sortDir param must be 'asc' or 'desc'"}
		}
		sortBy = sortMapping[sortBy] + " " + sortDir
	}

	adjustedPageNum := pageNumber - 1
	vendorItemListPage, vendorItemAmount, err := vendorService.vendorRepo.GetVendorItemListPage(c, uid, adjustedPageNum, sortBy, filterOn, searchStr, urlPath)
	if err != nil {
		return nil, err
	}

	// build batch url sign object for content service
	urlBatch := map[int]map[string]*string{}
	for _, vendorItem := range vendorItemListPage {
		if _, ok := urlBatch[int(*vendorItem.ItemId)]; !ok {
			urlBatch[int(*vendorItem.ItemId)] = map[string]*string{
				"contentMainUrl":  vendorItem.ContentMainUrl,
				"contentThumbUrl": vendorItem.ContentThumbUrl,
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

	earliestExpired := uint64(math.MaxUint64)
	for _, vendorItem := range vendorItemListPage {
		re := regexp.MustCompile(`Expires=(\d+)`)
		expiresValueStr := ""

		vendorItem.ContentMainUrl = signedUrlBatch[int(*vendorItem.ItemId)]["contentMainUrl"]
		vendorItem.ContentThumbUrl = signedUrlBatch[int(*vendorItem.ItemId)]["contentThumbUrl"]

		if vendorItem.ContentMainUrl != nil {
			if *vendorItem.ContentMainUrl != "" {
				match := re.FindStringSubmatch(*vendorItem.ContentMainUrl)
				if len(match) >= 2 {
					expiresValueStr = match[1]
				}
				expiresValue, err := strconv.ParseUint(expiresValueStr, 10, 64)
				if err != nil {
					return nil, err
				}
				if expiresValue < earliestExpired {
					earliestExpired = expiresValue
				}
				vendorItem.ContentExpiredTime = &expiresValueStr
			}
		}
	}

	pageSizeStr := os.Getenv("VENDOR_ITEM_LIST_PAGE_SIZE")
	pageSize, err := strconv.ParseUint(pageSizeStr, 10, 64)
	if err != nil {
		return nil, err
	}
	thisPageSize := uint64(len(vendorItemListPage))
	nextPageNum := uint64(0)
	nextPage := &nextPageNum
	if thisPageSize < pageSize {
		nextPage = nil
	} else {
		nextPageNum = (pageNumber + uint64(1))
	}
	result := model.VendorItemPage{
		VendorItemAmount: vendorItemAmount,
		EarliestExpired:  &earliestExpired,
		PageSize:         &thisPageSize,
		NextPage:         nextPage,
		Page:             vendorItemListPage,
	}
	return &result, nil
}

func (vendorService *VendorSvcImpl) GetActiveVendor(c context.Context, uid string) (*model.Vendor, error) {
	vendor, err := vendorService.vendorRepo.GetActiveVendor(c, uid)
	if err != nil {
		return nil, err
	}
	if vendor.UID == nil {
		return nil, &core.ErrorResp{
			Message: "creator is not currently active",
		}
	}
	return vendor, err
}

func (vendorService *VendorSvcImpl) GetVendor(c context.Context, uid string) (*model.Vendor, error) {
	return vendorService.vendorRepo.GetVendor(c, uid)
}

func (vendorService *VendorSvcImpl) GetVendorCategories(c context.Context, vendorId string) ([]*model.VendorCategoryExpanded, error) {
	return vendorService.vendorRepo.GetVendorCategories(c, vendorId)
}

func (vendorService *VendorSvcImpl) AddVendorCategories(c context.Context, categories []*model.VendorCategory) error {
	vendorId := categories[0].VendorId

	err := vendorService.vendorRepo.AddVendorCategories(c, categories)
	if err != nil {
		return err
	}

	// remove vendor categories cache
	if err := vendorService.ClearVendorCache(c, *vendorId); err != nil {
		return err
	}
	return nil
}

func (vendorService *VendorSvcImpl) ApproveVendor(c context.Context, uid string) (*model.Vendor, error) {
	// NOTE: not clearing the cache and new vendor will appear within T - one hour
	vendor, err := vendorService.vendorRepo.ApproveVendor(c, uid)
	if err != nil {
		return nil, err
	}
	return vendor, err
}

func (vendorService *VendorSvcImpl) RemoveVendor(c context.Context, uid string) error {
	return vendorService.vendorRepo.RemoveVendor(c, uid)
}

func (vendorService *VendorSvcImpl) PatchVendor(c context.Context, vendorId string, patchMap map[string]interface{}) (*model.Vendor, error) {
	dbPatchMap := core.ConvertJSONMapToDBMap(patchMap, model.User{})
	err := vendorService.vendorRepo.PatchVendor(c, vendorId, dbPatchMap)
	if err != nil {
		return nil, err
	}

	// remove vendor cache
	if err := vendorService.vendorRepo.ClearVendorCache(c, vendorId); err != nil {
		return nil, err
	}

	vendor, err := vendorService.GetActiveVendor(c, vendorId)
	if err != nil {
		return nil, err
	}
	return vendor, err
}

func (vendorService *VendorSvcImpl) ClearVendorCache(c context.Context, vendorId string) error {
	return vendorService.vendorRepo.ClearVendorCache(c, vendorId)
}
