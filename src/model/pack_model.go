package model

type Pack struct {
	ID              *uint64 `json:"id"`
	PackConfigId    *uint64 `json:"packConfigId"`
	PackItemQty     *uint64 `json:"packItemQty"`
	ImageUrl        *string `json:"imageUrl"`
	VendorId        *string `json:"vendorId"`
	CreatedAt       *string `json:"createdAt"`
	PurchasedAt     *string `json:"purchasedAt"`
	OpenedAt        *string `json:"openedAt"`
	OwnerId         *string `json:"ownerId"`
	Active          *bool   `json:"active"`
	Items           []*Item `json:"items"`
	Description     *string `json:"description"`
	Title           *string `json:"title"`
	ContentMainUrl  *string `json:"contentMainUrl"`
	ContentThumbUrl *string `json:"contentThumbUrl"`
}

type UserPack struct {
	PackConfigId      *uint64  `db:"pack_config_id" json:"packConfigId"`
	ImageUrl          *string  `db:"image_url" json:"imageUrl"`
	ItemQty           *uint64  `db:"item_qty" json:"itemQty"`
	Description       *string  `db:"description" json:"description"`
	Title             *string  `db:"title" json:"title"`
	VendorId          *string  `db:"vendor_id" json:"vendorId"`
	CreatedAt         *string  `db:"created_at" json:"createdAt"`
	TokenAmount       *float64 `db:"token_amount" json:"tokenAmount"`
	Qty               *uint64  `db:"qty" json:"qty"`
	RawCategories     *string  `db:"raw_categories" json:"rawCategories"`
	Categories        []string `db:"categories" json:"categories"`
	RawPackIds        *string  `db:"raw_pack_ids" json:"rawPackIds"`
	PackIds           []uint64 `db:"pack_ids" json:"packIds"`
	PackAmount        *uint64  `db:"pack_amount" json:"packAmount"`
	LatestPurchasedAt *string  `db:"latest_purchased_at" json:"latestPurchasedAt"`
	ContentMainUrl    *string  `db:"content_main_url" json:"cotentMainUrl"`
	ContentThumbUrl   *string  `db:"content_thumb_url" json:"contentThumbUrl"`
}

type PackFlatten struct {
	PackFactId          *uint64 `db:"pack_fact_id" json:"packFactId"`
	PackConfigId        *uint64 `db:"pack_config_id" json:"packConfigId"`
	ImageUrl            *string `db:"image_url" json:"imageUrl"`
	PackItemQty         *uint64 `db:"pack_item_qty" json:"PackItemQty"`
	VendorId            *string `db:"vendor_id" json:"vendorId"`
	CreatedAt           *string `db:"created_at" json:"createdAt"`
	PurchasedAt         *string `db:"purchased_at" json:"purchasedAt"`
	OpenedAt            *string `db:"opened_at" json:"openedAt"`
	OwnerId             *string `db:"owner_id" json:"ownerId"`
	Active              *bool   `db:"active" json:"active"`
	Description         *string `db:"description" json:"Description"`
	Title               *string `db:"title" json:"title"`
	ItemId              *uint64 `db:"item_id" json:"itemId"`
	ItemCreatedAt       *string `db:"item_created_at" json:"itemCreatedAt"`
	ItemDeletedAt       *string `db:"item_deleted_at" json:"itemDeletedAt"`
	ItemUpdatedAt       *string `db:"item_updated_at" json:"itemUpdatedAt"`
	ItemDescription     *string `db:"item_description" json:"itemDescription"`
	ItemVendorId        *string `db:"item_vendor_id" json:"itemVendorId"`
	ItemName            *string `db:"item_name" json:"itemName"`
	ItemImageUrl        *string `db:"item_image_url" json:"itemImageUrl"`
	ItemRarityId        *uint64 `db:"item_rarity_id" json:"itemRarityId"`
	ItemContentMainUrl  *string `db:"item_content_main_url" json:"itemContentMainUrl"`
	ItemContentThumbUrl *string `db:"item_content_thumb_url" json:"itemContentThumbUrl"`
	ItemContentType     *string `db:"item_content_type" json:"itemContentType"`
	ItemActive          *bool   `db:"item_active" json:"itemActive"`
	ItemNotify          *bool   `db:"item_notify" json:"itemNotify"`
}

type PackBoughtResp struct {
	PackIds    []uint64 `json:"packIds"`
	NewBalance float64  `json:"newBalance"`
}
