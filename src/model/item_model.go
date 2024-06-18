package model

type UserItemExpanded struct {
	ID          *uint64 `db:"id" json:"id"`
	VendorId    *string `db:"vendor_id" json:"vendorId"`
	ImageUrl    *string `db:"image_url" json:"imageUrl"`
	CreatedAt   *string `db:"created_at" json:"createdAt"`
	DeletedAt   *string `db:"deleted_at" json:"deletedAt"`
	UpdatedAt   *string `db:"updated_at" json:"updatedAt"`
	Description *string `db:"description" json:"description"`
	Name        *string `db:"name" json:"name"`
	Uid         *string `db:"uid" json:"uid"`
	Qty         *uint64 `db:"qty" json:"qty"`
	RemovedAt   *string `db:"removed_at" json:"removedAt"`
	AcquiredAt  *string `db:"acquired_at" json:"acquiredAt"`
}

type PackItemConfigExpanded struct {
	ID              *uint64  `db:"id" json:"id"`
	ItemQty         *uint64  `db:"item_qty" json:"itemQty"`
	ItemValue       *float64 `db:"item_value" json:"itemValue"`
	CreatedAt       *string  `db:"created_at" json:"createdAt"`
	DeletedAt       *string  `db:"deleted_at" json:"deletedAt"`
	UpdatedAt       *string  `db:"updated_at" json:"updatedAt"`
	Description     *string  `db:"description" json:"description"`
	VendorId        *string  `db:"vendor_id" json:"vendorId"`
	Name            *string  `db:"name" json:"name"`
	Rarity          *string  `db:"rarity" json:"rarity"`
	ContentMainUrl  *string  `db:"content_main_url" json:"contentMainUrl"`
	ContentThumbUrl *string  `db:"content_thumb_url" json:"contentThumbUrl"`
	ContentType     *string  `db:"content_type" json:"contentType"`
}
