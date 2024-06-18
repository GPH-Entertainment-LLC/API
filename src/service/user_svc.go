package service

import (
	"context"
	"fmt"
	"math"
	"os"
	"regexp"
	"slices"
	"strconv"
	"time"
	"xo-packs/core"
	"xo-packs/model"
	repository "xo-packs/repository"

	"github.com/gin-gonic/gin"
)

type UserService interface {
	CreateUser(context.Context, *model.User) (*model.User, error)
	GetUser(context.Context, string, bool) (*model.User, error)
	GetUserByUid(context.Context, string) (*model.User, error)
	GetUserByUsername(context.Context, string) (*model.User, error)
	GetUserItem(context.Context, uint64) (*model.UserItem, error)
	WithdrawalUserItem(context.Context, uint64, string) (*model.ItemWithdrawal, error)
	PatchUser(context.Context, string, string, map[string]interface{}) (*model.User, error)
	DeleteUser(context.Context, string, string) (*model.User, error)
	GetUserPackPage(context.Context, string, uint64, string, string, string, string, string) (*model.UserPackPage, error)
	GetUserItemPage(context.Context, string, uint64, string, string, string, string, string, ItemService) (*model.UserItemPage, error)
	GetUserFavoritesPage(*gin.Context, string, uint64, string, string) (*model.UserFavoritePage, error)
	ClearUserCache(context.Context, string, string) error
	ClearFavoriteCache(context.Context, string, string) error
	GetFavorite(context.Context, string, string) (*model.Favorite, error)
	AddFavorite(context.Context, string, string, VendorService) (*model.Favorite, error)
	RemoveFavorite(context.Context, string, string, VendorService) error
	FlushCache(context.Context) error
}

type UserSvcImpl struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	userService := UserSvcImpl{userRepo: userRepo}
	return &userService
}

func (userService *UserSvcImpl) CreateUser(c context.Context, user *model.User) (*model.User, error) {
	user, err := userService.userRepo.CreateUser(c, user)
	if err != nil {
		return nil, err
	}

	// clearing user cache
	if err := userService.ClearUserCache(c, *user.Uid, *user.Username); err != nil {
		return nil, err
	}

	// getting newly created user
	createdUser, err := userService.userRepo.GetUser(c, *user.Uid, true)
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

func (userService *UserSvcImpl) GetUser(c context.Context, uid string, private bool) (*model.User, error) {
	return userService.userRepo.GetUser(c, uid, private)
}

func (userService *UserSvcImpl) GetUserByUid(c context.Context, uid string) (*model.User, error) {
	return userService.userRepo.GetUser(c, uid, false)
}

func (userService *UserSvcImpl) GetUserByUsername(c context.Context, username string) (*model.User, error) {
	return userService.userRepo.GetUserByUsername(c, username)
}

func (userService *UserSvcImpl) DeleteUser(c context.Context, uid string, username string) (*model.User, error) {
	err := userService.userRepo.DeleteUser(c, uid, username)
	if err != nil {
		return nil, err
	}

	// clear user cache
	if err := userService.ClearUserCache(c, uid, username); err != nil {
		return nil, err
	}

	user, err := userService.GetUser(c, uid, false)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (userService *UserSvcImpl) PatchUser(c context.Context, uid string, username string, userPatchMap map[string]interface{}) (*model.User, error) {
	user, err := userService.GetUser(c, uid, true)
	if err != nil {
		return nil, err
	}

	fmt.Println("Last Username Changed: ", user.LastUsernameChangeAt)

	if len(userPatchMap) <= 0 {
		return user, nil
	}

	// checking if user has waited sufficient time before updating their username again
	_, exists := userPatchMap["username"]
	if exists {
		fmt.Println("Updating Username")
		currTime := time.Now()
		if user.LastUsernameChangeAt != nil {
			lastUsernameChange, err := time.Parse(time.RFC3339, *user.LastUsernameChangeAt)
			if err != nil {
				return nil, err
			}
			timeDifference := currTime.Sub(lastUsernameChange)
			rawMaxUsernameChangeDays := os.Getenv("MAX_USERNAME_CHANGE_DAYS")
			maxUsernameChangeDays, err := strconv.ParseFloat(rawMaxUsernameChangeDays, 10)
			if err != nil {
				return nil, err
			}
			daysLeft := int(maxUsernameChangeDays) - int((math.Floor(timeDifference.Hours()) / 24))
			if daysLeft > 0 {
				return nil, &core.ErrorResp{
					Message: fmt.Sprintf("You must wait %v days before changing your username again", daysLeft),
				}
			}
		}
		// new username change timestamp
		userPatchMap["last_username_change_at"] = currTime
	}

	dbPatchMap := core.ConvertJSONMapToDBMap(userPatchMap, model.User{})
	err = userService.userRepo.PatchUser(c, uid, username, dbPatchMap)
	if err != nil {
		return nil, err
	}

	// flush vendor cache
	if *user.IsVendor == true {
		if err = userService.FlushCache(c); err != nil {
			return nil, err
		}
	} else {
		if err = userService.ClearUserCache(c, uid, username); err != nil {
			return nil, err
		}
	}

	user, err = userService.GetUser(c, uid, true)
	if err != nil {
		return nil, err
	}
	return user, err
}

func (userService *UserSvcImpl) GetUserItem(c context.Context, userItemId uint64) (*model.UserItem, error) {
	return userService.userRepo.GetUserItem(c, userItemId)
}

func (userService *UserSvcImpl) WithdrawalUserItem(c context.Context, userItemId uint64, uid string) (*model.ItemWithdrawal, error) {
	userItem, err := userService.GetUserItem(c, userItemId)
	if err != nil {
		return nil, err
	}

	// validate authorized user is the owner of item
	if userItem.Uid == nil || (userItem.Uid != nil && *userItem.Uid != uid) {
		return nil, &core.ErrorResp{
			Message: "user item does not exist for authorized user",
		}
	}

	// create new withdrawal for user item
	withdrawalId, err := userService.userRepo.WithdrawalUserItem(c, userItemId)
	if err != nil {
		return nil, err
	}
	if withdrawalId == nil {
		return nil, &core.ErrorResp{
			Message: "unable to withdrawal item at this time",
		}
	}

	// get newly created withdrawal
	itemWithdrawal, err := userService.userRepo.GetUserItemWithdrawal(c, *withdrawalId)
	if err != nil {
		return nil, err
	}
	return itemWithdrawal, nil
}

func (userService *UserSvcImpl) GetUserPackPage(
	c context.Context, uid string, pageNumber uint64, sortBy string, sortDir string, filterOn string, searchStr string, urlPath string) (*model.UserPackPage, error) {

	if filterOn != "" {
		categories, err := userService.userRepo.GetUserPackCategories(c)
		if err != nil {
			return nil, err
		}
		if !slices.Contains(categories, filterOn) {
			return nil, &core.ErrorResp{Message: "Filter on category not valid"}
		}
	}
	if sortBy != "" {
		sortMapping, err := userService.userRepo.GetUserPackSortMappings(c)
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
	userPackPage, userPackAmount, err := userService.userRepo.GetUserPackPage(c, uid, adjustedPageNum, sortBy, filterOn, searchStr, urlPath)
	if err != nil {
		return nil, err
	}
	pageSizeStr := os.Getenv("USER_PACKS_PAGE_SIZE")
	pageSize, err := strconv.ParseUint(pageSizeStr, 10, 64)
	if err != nil {
		return nil, err
	}
	thisPageSize := uint64(len(userPackPage))
	nextPageNum := uint64(0)
	nextPage := &nextPageNum
	if thisPageSize < pageSize {
		nextPage = nil
	} else {
		nextPageNum = (pageNumber + uint64(1))
	}
	result := model.UserPackPage{UserPackAmount: userPackAmount, PageSize: &thisPageSize, NextPage: nextPage, Page: userPackPage}
	return &result, nil
}

func (userService *UserSvcImpl) GetUserItemPage(
	c context.Context, uid string, pageNumber uint64, sortBy string, sortDir string, filterOn string, searchStr string, urlPath string, itemService ItemService) (*model.UserItemPage, error) {

	if filterOn != "" {
		categories, err := userService.userRepo.GetUserItemCategories(c)
		if err != nil {
			return nil, err
		}
		if !slices.Contains(categories, filterOn) {
			return nil, &core.ErrorResp{Message: "Filter on category not valid"}
		}
	}
	if sortBy != "" {
		sortMapping, err := userService.userRepo.GetUserItemSortMappings(c)
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
	userItemPage, userItemAmount, err := userService.userRepo.GetUserItemPage(c, uid, adjustedPageNum, sortBy, filterOn, searchStr, urlPath)
	if err != nil {
		return nil, err
	}

	// build batch url sign object for content service
	urlBatch := map[int]map[string]*string{}
	for _, userItem := range userItemPage {
		if _, ok := urlBatch[int(*userItem.ItemId)]; !ok {
			urlBatch[int(*userItem.ItemId)] = map[string]*string{
				"contentMainUrl":  userItem.ContentMainUrl,
				"contentThumbUrl": userItem.ContentThumbUrl,
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
	for _, userItem := range userItemPage {
		re := regexp.MustCompile(`Expires=(\d+)`)
		expiresValueStr := ""

		userItem.ContentMainUrl = signedUrlBatch[int(*userItem.ItemId)]["contentMainUrl"]
		userItem.ContentThumbUrl = signedUrlBatch[int(*userItem.ItemId)]["contentThumbUrl"]

		if userItem.ContentMainUrl != nil {
			if *userItem.ContentMainUrl != "" {
				match := re.FindStringSubmatch(*userItem.ContentMainUrl)
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
				userItem.ContentExpiredTime = &expiresValueStr
			}
		}
	}

	pageSizeStr := os.Getenv("USER_ITEMS_PAGE_SIZE")
	pageSize, err := strconv.ParseUint(pageSizeStr, 10, 64)
	if err != nil {
		return nil, err
	}
	thisPageSize := uint64(len(userItemPage))
	nextPageNum := uint64(0)
	nextPage := &nextPageNum
	if thisPageSize < pageSize {
		nextPage = nil
	} else {
		nextPageNum = (pageNumber + uint64(1))
	}
	result := model.UserItemPage{
		UserItemAmount:  userItemAmount,
		EarliestExpired: &earliestExpired,
		PageSize:        &thisPageSize,
		NextPage:        nextPage,
		Page:            userItemPage,
	}
	return &result, nil
}

func (userService *UserSvcImpl) GetUserFavoritesPage(c *gin.Context, uid string, pageNumber uint64, searchStr string, urlPath string) (*model.UserFavoritePage, error) {
	adjustedPageNum := pageNumber - 1
	userFavoritePage, userFavoriteAmount, err := userService.userRepo.GetUserFavoritesPage(c, uid, adjustedPageNum, searchStr, urlPath)
	if err != nil {
		return nil, err
	}
	pageSizeStr := os.Getenv("USER_FAVORITES_PAGE_SIZE")
	pageSize, err := strconv.ParseUint(pageSizeStr, 10, 64)
	if err != nil {
		return nil, err
	}
	thisPageSize := uint64(len(userFavoritePage))
	nextPageNum := uint64(0)
	nextPage := &nextPageNum
	if thisPageSize < pageSize {
		nextPage = nil
	} else {
		nextPageNum = (pageNumber + uint64(1))
	}
	result := model.UserFavoritePage{UserFavoriteAmount: userFavoriteAmount, PageSize: &thisPageSize, NextPage: nextPage, Page: userFavoritePage}
	return &result, nil
}

func (userService *UserSvcImpl) ClearUserCache(c context.Context, uid string, username string) error {
	return userService.userRepo.ClearUserCache(c, uid, username)
}

func (userService *UserSvcImpl) ClearFavoriteCache(c context.Context, uid string, vendorId string) error {
	return userService.userRepo.ClearFavoriteCache(c, uid, vendorId)
}

func (userService *UserSvcImpl) GetFavorite(c context.Context, uid string, vendorId string) (*model.Favorite, error) {
	return userService.userRepo.GetFavorite(c, uid, vendorId)
}

func (userService *UserSvcImpl) AddFavorite(c context.Context, uid string, vendorId string, vendorService VendorService) (*model.Favorite, error) {
	favorite, err := userService.userRepo.AddFavorite(c, uid, vendorId)
	if err != nil {
		return nil, err
	}

	// clear vendor cache
	err = vendorService.ClearVendorCache(c, vendorId)
	if err != nil {
		return nil, err
	}

	// clear favorite cache
	err = userService.ClearFavoriteCache(c, uid, vendorId)
	if err != nil {
		return nil, err
	}
	return favorite, err
}

func (userService *UserSvcImpl) RemoveFavorite(c context.Context, uid string, vendorId string, vendorService VendorService) error {
	err := userService.userRepo.RemoveFavorite(c, uid, vendorId)
	if err != nil {
		return err
	}

	// clear vendor cache
	err = vendorService.ClearVendorCache(c, vendorId)
	if err != nil {
		return err
	}

	// clear favorite cache
	err = userService.ClearFavoriteCache(c, uid, vendorId)
	if err != nil {
		return err
	}
	return nil
}

func (userService *UserSvcImpl) FlushCache(c context.Context) error {
	return userService.userRepo.FlushCache(c)
}
