package model

type PageVendor struct {
	UID             *string  `db:"uid" json:"uid" `
	CreatedAt       *string  `db:"created_at" json:"createdAt"`
	UpdatedAt       *string  `db:"updated_at" json:"updatedAt"`
	ImageUrl        *string  `db:"image_url" json:"imageUrl"`
	FirstName       *string  `db:"first_name" json:"firstName"`
	LastName        *string  `db:"last_name" json:"lastName"`
	DisplayName     *string  `db:"display_name" json:"displayName"`
	Username        *string  `db:"username" json:"username"`
	Birthday        *string  `db:"birthday" json:"birthday"`
	Bio             *string  `db:"bio" json:"bio"`
	IsVendor        *bool    `db:"is_vendor" json:"isVendor"`
	Active          *bool    `db:"active" json:"active"`
	PackAmount      *uint64  `db:"pack_amount" json:"packAmount"`
	FavoritesAmount *uint64  `db:"favorites_amount" json:"favoritesAmount"`
	RawCategories   *string  `db:"categories" json:"rawCategories"`
	Categories      []string `json:"categories"`
	VendorAmount    *uint64  `db:"vendor_amount" json:"vendorAmount"`
	ContentMainUrl  *string  `db:"content_main_url" json:"contentMainUrl"`
	ContentThumbUrl *string  `db:"content_thumb_url" json:"contentThumbUrl"`
	BannerUrl       *string  `db:"banner_url" json:"bannerUrl"`
}

type PageUserPack struct {
	PackConfigId          *uint64  `db:"pack_config_id" json:"packConfigId"`
	ImageUrl              *string  `db:"image_url" json:"imageUrl"`
	ItemQty               *uint64  `db:"item_qty" json:"itemQty"`
	Description           *string  `db:"description" json:"description"`
	Title                 *string  `db:"title" json:"title"`
	CreatedAt             *string  `db:"created_at" json:"createdAt"`
	TokenAmount           *float64 `db:"token_amount" json:"tokenAmount"`
	Qty                   *uint64  `db:"qty" json:"qty"`
	ContentMainUrl        *string  `db:"content_main_url" json:"contentMainUrl"`
	ContentThumbUrl       *string  `db:"content_thumb_url" json:"contentThumbUrl"`
	RawCategories         *string  `db:"raw_categories" json:"rawCategories"`
	Categories            []string `db:"categories" json:"categories"`
	RawPackIds            *string  `db:"raw_pack_ids" json:"rawPackIds"`
	PackIds               []uint64 `db:"pack_ids" json:"packIds"`
	VendorId              *string  `db:"vendor_id" json:"vendorId"`
	VendorFirstName       *string  `db:"vendor_first_name" json:"vendorFirstName"`
	VendorLastName        *string  `db:"vendor_last_name" json:"vendorLastName"`
	VendorUsername        *string  `db:"vendor_username" json:"vendorUsername"`
	VendorDisplayName     *string  `db:"vendor_display_name" json:"vendorDisplayName"`
	VendorImageUrl        *string  `db:"vendor_image_url" json:"vendorImageUrl"`
	VendorFavoritesAmount *uint64  `db:"vendor_favorites_amount" json:"vendorFavoritesAmount"`
	VendorBio             *string  `db:"vendor_bio" json:"vendorBio"`
	VendorActive          *bool    `db:"vendor_active" json:"vendorActive"`
	VendorMainUrl         *string  `db:"vendor_main_url" json:"vendorMainUrl"`
	VendorThumbUrl        *string  `db:"vendor_thumb_url" json:"vendorThumbUrl"`
	VendorBannerUrl       *string  `db:"vendor_banner_url" json:"vendorBannerUrl"`
	TotalPacksAmount      *uint64  `db:"total_packs_amount" json:"totalPacksAmount"`
	LatestPurchasedAt     *string  `db:"latest_purchased_at" json:"latestPurchasedAt"`
	PackAmount            *uint64  `db:"pack_amount" json:"packAmount"`
}

type PageVendorPack struct {
	PackConfigId          *uint64  `db:"pack_config_id" json:"packConfigId"`
	ImageUrl              *string  `db:"image_url" json:"imageUrl"`
	ItemQty               *uint64  `db:"item_qty" json:"itemQty"`
	Description           *string  `db:"description" json:"description"`
	Title                 *string  `db:"title" json:"title"`
	Active                *bool    `db:"active" json:"active"`
	CreatedAt             *string  `db:"created_at" json:"createdAt"`
	UpdatedAt             *string  `db:"updated_at" json:"updatedAt"`
	DeletedAt             *string  `db:"deleted_at" json:"deletedAt"`
	ReleaseAt             *string  `db:"release_at" json:"releaseAt"`
	TokenAmount           *float64 `db:"token_amount" json:"tokenAmount"`
	Qty                   *uint64  `db:"qty" json:"qty"`
	ContentMainUrl        *string  `db:"content_main_url" json:"contentMainUrl"`
	ContentThumbUrl       *string  `db:"content_thumb_url" json:"contentThumbUrl"`
	RawCategories         *string  `db:"raw_categories" json:"rawCategories"`
	Categories            []string `db:"categories" json:"categories"`
	CurrentStock          *uint64  `db:"current_stock" json:"currentStock"`
	QtySold               *uint64  `db:"qty_sold" json:"qtySold"`
	VendorId              *string  `db:"vendor_id" json:"vendorId"`
	VendorFirstName       *string  `db:"vendor_first_name" json:"vendorFirstName"`
	VendorLastName        *string  `db:"vendor_last_name" json:"vendorLastName"`
	VendorUsername        *string  `db:"vendor_username" json:"vendorUsername"`
	VendorDisplayName     *string  `db:"vendor_display_name" json:"vendorDisplayName"`
	VendorImageUrl        *string  `db:"vendor_image_url" json:"vendorImageUrl"`
	VendorFavoritesAmount *uint64  `db:"vendor_favorites_amount" json:"vendorFavoritesAmount"`
	VendorBio             *string  `db:"vendor_bio" json:"vendorBio"`
	VendorActive          *bool    `db:"vendor_active" json:"vendorActive"`
	VendorMainUrl         *string  `db:"vendor_main_url" json:"vendorMainUrl"`
	VendorThumbUrl        *string  `db:"vendor_thumb_url" json:"vendorThumbUrl"`
	VendorBannerUrl       *string  `db:"vendor_banner_url" json:"vendorBannerUrl"`
	TotalPacksAmount      *uint64  `db:"total_packs_amount" json:"totalPacksAmount"`
}

type PageUserItem struct {
	ItemId                *uint64  `db:"item_id" json:"itemId"`
	CreatedAt             *string  `db:"created_at" json:"createdAt"`
	DeletedAt             *string  `db:"deleted_at" json:"deletedAt"`
	UpdatedAt             *string  `db:"updated_at" json:"updatedAt"`
	Description           *string  `db:"description" json:"description"`
	Name                  *string  `db:"name" json:"name"`
	ImageUrl              *string  `db:"image_url" json:"imageUrl"`
	RawCategories         *string  `db:"raw_categories" json:"rawCategories"`
	Categories            []string `db:"categories" json:"categories"`
	Rarity                *string  `db:"rarity" json:"rarity"`
	Ranking               *uint64  `db:"ranking" json:"ranking"`
	ContentMainUrl        *string  `db:"content_main_url" json:"contentMainUrl"`
	ContentThumbUrl       *string  `db:"content_thumb_url" json:"contentThumbUrl"`
	ContentExpiredTime    *string  `db:"content_expired_time" json:"contentExpiredTime"`
	ContentType           *string  `db:"content_type" json:"contentType"`
	Active                *bool    `db:"active" json:"active"`
	Notify                *bool    `db:"notify" json:"notify"`
	Value                 *float64 `db:"value" json:"value"`
	VendorId              *string  `db:"vendor_id" json:"vendorId"`
	VendorFirstName       *string  `db:"vendor_first_name" json:"vendorFirstName"`
	VendorLastName        *string  `db:"vendor_last_name" json:"vendorLastName"`
	VendorUsername        *string  `db:"vendor_username" json:"vendorUsername"`
	VendorDisplayName     *string  `db:"vendor_display_name" json:"vendorDisplayName"`
	VendorImageUrl        *string  `db:"vendor_image_url" json:"vendorImageUrl"`
	VendorFavoritesAmount *uint64  `db:"vendor_favorites_amount" json:"vendorFavoritesAmount"`
	VendorBio             *string  `db:"vendor_bio" json:"vendorBio"`
	VendorActive          *bool    `db:"vendor_active" json:"vendorActive"`
	VendorMainUrl         *string  `db:"vendor_main_url" json:"vendorMainUrl"`
	VendorThumbUrl        *string  `db:"vendor_thumb_url" json:"vendorThumbUrl"`
	VendorBannerUrl       *string  `db:"vendor_banner_url" json:"vendorBannerUrl"`
	UserItemId            *uint64  `db:"user_item_id" json:"userItemId"`
	TotalItemsAmount      *uint64  `db:"total_items_amount" json:"totalItemsAmount"`
	OwnerId               *string  `db:"owner_id" json:"ownerId"`
	AcquiredAt            *string  `db:"acquired_at" json:"acquiredAt"`
	ExpiredAt             *string  `db:"expiredAt" json:"expiredAt"`
}

type PageVendorItem struct {
	ItemId                *uint64  `db:"item_id" json:"itemId"`
	CreatedAt             *string  `db:"created_at" json:"createdAt"`
	DeletedAt             *string  `db:"deleted_at" json:"deletedAt"`
	UpdatedAt             *string  `db:"updated_at" json:"updatedAt"`
	Description           *string  `db:"description" json:"description"`
	Name                  *string  `db:"name" json:"name"`
	ImageUrl              *string  `db:"image_url" json:"imageUrl"`
	RawCategories         *string  `db:"raw_categories" json:"rawCategories"`
	Categories            []string `db:"categories" json:"categories"`
	Rarity                *string  `db:"rarity" json:"rarity"`
	Ranking               *uint64  `db:"ranking" json:"ranking"`
	ContentMainUrl        *string  `db:"content_main_url" json:"contentMainUrl"`
	ContentThumbUrl       *string  `db:"content_thumb_url" json:"contentThumbUrl"`
	ContentExpiredTime    *string  `db:"content_expired_time" json:"contentExpiredTime"`
	ContentType           *string  `db:"content_type" json:"contentType"`
	Active                *bool    `db:"active" json:"active"`
	Notify                *bool    `db:"notify" json:"notify"`
	Value                 *float64 `db:"value" json:"value"`
	VendorId              *string  `db:"vendor_id" json:"vendorId"`
	VendorFirstName       *string  `db:"vendor_first_name" json:"vendorFirstName"`
	VendorLastName        *string  `db:"vendor_last_name" json:"vendorLastName"`
	VendorUsername        *string  `db:"vendor_username" json:"vendorUsername"`
	VendorDisplayName     *string  `db:"vendor_display_name" json:"vendorDisplayName"`
	VendorImageUrl        *string  `db:"vendor_image_url" json:"vendorImageUrl"`
	VendorFavoritesAmount *uint64  `db:"vendor_favorites_amount" json:"vendorFavoritesAmount"`
	VendorBio             *string  `db:"vendor_bio" json:"vendorBio"`
	VendorActive          *bool    `db:"vendor_active" json:"vendorActive"`
	VendorMainUrl         *string  `db:"vendor_main_url" json:"vendorMainUrl"`
	VendorThumbUrl        *string  `db:"vendor_thumb_url" json:"vendorThumbUrl"`
	VendorBannerUrl       *string  `db:"vendor_banner_url" json:"vendorBannerUrl"`
	TotalItemsAmount      *uint64  `db:"total_items_amount" json:"totalItemsAmount"`
}

type PageUserFavorite struct {
	FavoriteId            *uint64 `db:"favorite_id" json:"favoriteId"`
	FavoritedAt           *string `db:"favorited_at" json:"favoritedAt"`
	Uid                   *string `db:"uid" json:"uid"`
	VendorId              *string `db:"vendor_id" json:"vendorId"`
	VendorCreatedAt       *string `db:"vendor_created_at" json:"vendorCreatedAt"`
	VendorUpdatedAt       *string `db:"vendor_updated_at" json:"vendorUpdatedAt"`
	VendorFirstName       *string `db:"vendor_first_name" json:"vendorFirstName"`
	VendorLastName        *string `db:"vendor_last_name" json:"vendorLastName"`
	VendorUsername        *string `db:"vendor_username" json:"vendorUsername"`
	VendorBirthday        *string `db:"vendor_birthday" json:"vendorBirthday"`
	VendorBio             *string `db:"vendor_bio" json:"vendorBio"`
	VendorActive          *bool   `db:"vendor_active" json:"vendorActive"`
	VendorDisplayName     *string `db:"vendor_display_name" json:"vendorDisplayName"`
	VendorFavoritesAmount *uint64 `db:"vendor_favorites_amount" json:"vendorFavoritesAmount"`
	VendorThumbUrl        *string `db:"vendor_thumb_url" json:"vendorThumbUrl"`
	VendorMainUrl         *string `db:"vendor_main_url" json:"vendorMainUrl"`
	VendorBannerUrl       *string `db:"vendor_banner_url" json:"vendorBannerUrl"`
	VendorPackAmount      *uint64 `db:"vendor_pack_amount" json:"vendorPackAmount"`
	FavoritesAmount       *uint64 `db:"favorites_amount" json:"favoritesAmount"`
}

type PageUserTransactionHistory struct {
	Uid                  *string  `db:"uid" json:"uid"`
	TokenBundleId        *int     `db:"token_bundle_id" json:"tokenBundleId"`
	SubscriptionId       *string  `db:"subscription_id" json:"subsccriptionId"`
	TransactionId        *string  `db:"transaction_id" json:"transactionId"`
	TranDateTime         *string  `db:"tran_datetime" json:"tranDateTime"`
	ClientAccnum         *string  `db:"client_accnum" json:"clientAccnum"`
	ClientSubacc         *string  `db:"client_subacc" json:"clientSubacc"`
	BilledInitialPrice   *string  `db:"billed_initial_price" json:"billedInitialPrice"`
	InitialPeriod        *string  `db:"initial_period" json:"initialPeriod"`
	BilledRecurringPrice *string  `db:"billed_recurring_price" json:"billedRecurringPrice"`
	RecurringPeriod      *string  `db:"recurring_period" json:"recurringPeriod"`
	Rebills              *string  `db:"rebills" json:"rebills"`
	CurrencyCode         *string  `db:"currency_code" json:"currencyCode"`
	DollarAmount         *float64 `db:"dollar_amount" json:"dollarAmount"`
	TokenAmount          *float64 `db:"token_amount" json:"tokenAmount"`
	BundleImageUrl       *string  `db:"bundle_image_url" json:"bundleImageUrl"`
	TransactionCount     *uint64  `db:"transaction_count" json:"transactionCount"`
}

type UserTransactionHistoryPage struct {
	TransactionAmount *uint64                       `json:"transactionAmount"`
	NextPage          *uint64                       `json:"nextPage"`
	PageSize          *uint64                       `json:"pageSize"`
	Page              []*PageUserTransactionHistory `json:"page"`
}

type VendorPage struct {
	VendorAmount *uint64       `json:"vendorAmount"`
	NextPage     *uint64       `json:"nextPage"`
	PageSize     *uint64       `json:"pageSize"`
	Page         []*PageVendor `json:"page"`
}

type VendorItemPage struct {
	VendorItemAmount *uint64           `json:"vendorItemAmount"`
	EarliestExpired  *uint64           `json:"earliestExpired"`
	NextPage         *uint64           `json:"nextPage"`
	PageSize         *uint64           `json:"pageSize"`
	Page             []*PageVendorItem `json:"page"`
}

type UserItemPage struct {
	UserItemAmount  *uint64         `json:"userItemAmount"`
	EarliestExpired *uint64         `json:"earliestExpired"`
	NextPage        *uint64         `json:"nextPage"`
	PageSize        *uint64         `json:"pageSize"`
	Page            []*PageUserItem `json:"page"`
}

type VendorPackPage struct {
	VendorPackAmount *uint64           `json:"vendorPackAmount"`
	NextPage         *uint64           `json:"nextPage"`
	PageSize         *uint64           `json:"pageSize"`
	Page             []*PageVendorPack `json:"page"`
}

type UserPackPage struct {
	UserPackAmount *uint64         `json:"userPackAmount"`
	NextPage       *uint64         `json:"nextPage"`
	PageSize       *uint64         `json:"pageSize"`
	Page           []*PageUserPack `json:"page"`
}

type UserFavoritePage struct {
	UserFavoriteAmount *uint64             `json:"favoriteAmount"`
	NextPage           *uint64             `json:"nextPage"`
	PageSize           *uint64             `json:"pageSize"`
	Page               []*PageUserFavorite `json:"page"`
}
