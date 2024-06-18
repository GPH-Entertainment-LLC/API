package model

type User struct {
	Uid                  *string `db:"uid" json:"uid" `
	Email                *string `db:"email" json:"email"`
	CreatedAt            *string `db:"created_at" json:"createdAt"`
	UpdatedAt            *string `db:"updated_at" json:"updatedAt"`
	DeletedAt            *string `db:"deleted_at" json:"deletedAt"`
	ImageUrl             *string `db:"image_url" json:"imageUrl"`
	FirstName            *string `db:"first_name" json:"firstName"`
	LastName             *string `db:"last_name" json:"lastName"`
	DisplayName          *string `db:"display_name" json:"displayName"`
	Username             *string `db:"username" json:"username"`
	Birthday             *string `db:"birthday" json:"birthday"`
	Bio                  *string `db:"bio" json:"bio"`
	IsVendor             *bool   `db:"is_vendor" json:"isVendor"`
	Active               *bool   `db:"active" json:"active"`
	ContenMainUrl        *string `db:"content_main_url" json:"contentMainUrl"`
	ContentThumbUrl      *string `db:"content_thumb_url" json:"contentThumbUrl"`
	BannerUrl            *string `db:"banner_url" json:"bannerUrl"`
	ProfileImgVersion    *int    `db:"profile_img_version" json:"profileImgVersion"`
	BannerImgVersion     *int    `db:"banner_img_version" json:"bannerImgVersion"`
	LastUsernameChangeAt *string `db:"last_username_change_at" json:"lastUsernameChangeAt"`
}

type PackConfig struct {
	ID              *uint64  `db:"id" json:"id"`
	ImageUrl        *string  `db:"image_url" json:"imageUrl"`
	VendorID        *string  `db:"vendor_id" json:"vendorId"`
	CreatedAt       *string  `db:"created_at" json:"createdAt"`
	DeletedAt       *string  `db:"deleted_at" json:"deletedAt"`
	UpdatedAt       *string  `db:"updated_at" json:"updatedAt"`
	ReleaseAt       *string  `db:"release_at" json:"releaseAt"`
	Description     *string  `db:"description" json:"description"`
	Title           *string  `db:"title" json:"title"`
	TokenAmount     *float64 `db:"token_amount" json:"tokenAmount"`
	Qty             *int     `db:"qty" json:"qty"`
	ItemQty         *int     `db:"item_qty" json:"itemQty"`
	CurrentStock    *int     `db:"current_stock" json:"currentStock"`
	ContentMainUrl  *string  `db:"content_main_url" json:"contentMainUrl"`
	ContentThumbUrl *string  `db:"content_thumb_url" json:"contentThumbUrl"`
	Active          *bool    `db:"active" json:"active"`
	QtySold         *uint64  `db:"qty_sold" json:"qtySold"`
}

type PackFact struct {
	ID           *uint64 `db:"id" json:"id"`
	CreatedAt    *string `db:"created_at" json:"createdAt"`
	PurchasedAt  *string `db:"purchased_at" json:"purchasedAt"`
	OpenedAt     *string `db:"opened_at" json:"openedAt"`
	OwnerID      *string `db:"owner_id" json:"ownerId"`
	PackConfigID *uint64 `db:"pack_config_id" json:"packConfigId"`
	Active       *bool   `db:"active" json:"active"`
}

type Item struct {
	ID              *uint64  `db:"id" json:"id"`
	VendorId        *string  `db:"vendor_id" json:"vendorId"`
	ImageUrl        *string  `db:"image_url" json:"imageUrl"`
	CreatedAt       *string  `db:"created_at" json:"createdAt"`
	DeletedAt       *string  `db:"deleted_at" json:"deletedAt"`
	UpdatedAt       *string  `db:"updated_at" json:"updatedAt"`
	Description     *string  `db:"description" json:"description"`
	Name            *string  `db:"name" json:"name"`
	RarityId        *uint64  `db:"rarity_id" json:"rarityId"`
	ContentMainUrl  *string  `db:"content_main_url" json:"contentMainUrl"`
	ContentThumbUrl *string  `db:"content_thumb_url" json:"contentThumbUrl"`
	ContentType     *string  `db:"content_type" json:"contentType"`
	Active          *bool    `db:"active" json:"active"`
	Notify          *bool    `db:"notify" json:"notify"`
	Value           *float64 `db:"value" json:"value"`
}

type PackItemConfig struct {
	ID           *uint64 `db:"id" json:"id"`
	PackConfigID *uint64 `db:"pack_config_id" json:"packConfigId"`
	ItemID       *uint64 `db:"item_id" json:"itemId"`
	Qty          *int    `db:"qty" json:"qty"`
	CreatedAt    *string `db:"created_at" json:"createdAt"`
	RemovedAt    *string `db:"removed_at" json:"removedAt"`
}

type PackItemFact struct {
	ID     *uint64 `db:"id" json:"id"`
	ItemID *uint64 `db:"item_id" json:"itemId"`
	PackID *uint64 `db:"pack_id" json:"packId"`
}

type PackItemPreview struct {
	PackConfigID    *uint64  `db:"pack_config_id" json:"packConfigId"`
	ItemID          *uint64  `db:"item_id" json:"id"`
	ItemName        *string  `db:"item_name" json:"itemName"`
	ItemContentType *string  `db:"item_content_type" json:"itemContentType"`
	ItemRarity      *string  `db:"item_rarity" json:"itemRarity"`
	ItemChance      *float64 `db:"item_chance" json:"itemChance"`
}

type UserItem struct {
	ID         *uint64 `db:"id" json:"id"`
	Uid        *string `db:"uid" json:"uid"`
	ItemId     *uint64 `db:"item_id" json:"itemId"`
	AcquiredAt *string `db:"acquired_at" json:"acquiredAt"`
	RemovedAt  *string `db:"removed_at" json:"removedAt"`
	ExpiredAt  *string `db:"expired_at" json:"expiredAt"`
}

type Favorite struct {
	ID          *uint64 `db:"id" json:"id"`
	Uid         *string `db:"uid" json:"uid"`
	VendorId    *string `db:"vendor_id" json:"vendorId"`
	FavoritedAt *string `db:"favorited_at" json:"favoritedAt"`
}

type Category struct {
	ID       *uint64 `db:"id" json:"id"`
	Category *string `db:"category" json:"category"`
}

type PackCategory struct {
	PackConfigId *uint64 `db:"pack_config_id" json:"packConfigId"`
	CategoryId   *uint64 `db:"category_id" json:"categoryId"`
}

type ItemCategory struct {
	ItemId     *uint64 `db:"item_id" json:"itemId"`
	CategoryId *uint64 `db:"category_id" json:"categoryId"`
}

type VendorCategory struct {
	VendorId   *string `db:"vendor_id" json:"vendorId"`
	CategoryId *uint64 `db:"category_id" json:"categoryId"`
}

type VendorCategoryExpanded struct {
	VendorId   *string `db:"vendor_id" json:"vendorId"`
	CategoryId *uint64 `db:"category_id" json:"categoryId"`
	Category   *string `db:"category" json:"category"`
}

type VendorApplication struct {
	ID            *uint64 `db:"id" json:"id"`
	Uid           *string `db:"uid" json:"uid"`
	Username      *string `db:"username" json:"username"`
	Email         *string `db:"email" json:"email"`
	FirstName     *string `db:"first_name" json:"firstName"`
	LastName      *string `db:"last_name" json:"lastName"`
	Birthday      *string `db:"birthday" json:"birthday"`
	StreetAddress *string `db:"street_address" json:"streetAddress"`
	AptNumber     *string `db:"apt_number" json:"aptNumber"`
	City          *string `db:"city" json:"city"`
	State         *string `db:"state" json:"state"`
	ZipCode       *string `db:"zip_code" json:"zipCode"`
	SubmittedAt   *string `db:"submitted_at" json:"submittedAt"`
	ApprovedAt    *string `db:"approved_at" json:"approvedAt"`
	RejectedAt    *string `db:"rejected_at" json:"rejectedAt"`
	Status        *string `db:"status" json:"status"`
	SocialLink1   *string `db:"social_link1" json:"socialLink1"`
	SocialLink2   *string `db:"social_link2" json:"socialLink2"`
	SocialLink3   *string `db:"social_link3" json:"socialLink3"`
	ReferralCode  *string `db:"referral_code" json:"referralCode"`
	PhoneNumber   *string `db:"phone_number" json:"phoneNumber"`
	Notes         *string `db:"notes" json:"notes"`
}

type ReferralCode struct {
	Code      *string `db:"code" json:"code"`
	Uid       *string `db:"uid" json:"uid"`
	Active    *bool   `db:"active" json:"active"`
	CreatedAt *string `db:"created_at" json:"createdAt"`
	RemovedAt *string `db:"removed_at" json:"removedAt"`
}

type Referral struct {
	ID          *uint64 `db:"id" json:"id"`
	RefereeUid  *string `db:"referee_uid" json:"refereeUid"`
	ReferrerUid *string `db:"referrer_uid" json:"referrerUid"`
	Code        *string `db:"code" json:"code"`
	ValidatedAt *string `db:"validated_at" json:"validatedAt"`
}

type ReportOpt struct {
	ID        *uint64 `db:"id" json:"id"`
	Opt       *string `db:"opt" json:"opt"`
	Active    *bool   `db:"active" json:"active"`
	CreatedAt *string `db:"created_at" json:"createdAt"`
	RemovedAt *string `db:"removed_at" json:"removedAt"`
}

type Report struct {
	ID          *uint64 `db:"id" json:"id"`
	Uid         *string `db:"uid" json:"uid"`
	ReportedUid *string `db:"reported_uid" json:"reportedUid"`
	OptId       *uint64 `db:"opt_id" json:"optId"`
	Notes       *string `db:"notes" json:"notes"`
	ReportedAt  *string `db:"reported_at" json:"reportedAt"`
	Active      *bool   `db:"active" json:"active"`
	ResolvedAt  *string `db:"resolved_at" json:"resolvedAt"`
}

type ReportExpanded struct {
	ReportId    *uint64 `db:"report_id" json:"reportId"`
	Uid         *string `db:"uid" json:"uid"`
	ReportedUid *string `db:"reported_uid" json:"reportedUid"`
	ReportedAt  *string `db:"reported_at" json:"reportedAt"`
	Opt         *string `db:"opt" json:"opt"`
	Notes       *string `db:"notes" json:"notes"`
	Active      *bool   `db:"active" json:"active"`
	ResolvedAt  *string `db:"resolved_at" json:"resolvedAt"`
}

type Faq struct {
	ID        *uint64 `db:"id" json:"id"`
	Question  *string `db:"question" json:"question"`
	Answer    *string `db:"answer" json:"answer"`
	Active    *bool   `db:"active" json:"active"`
	CreatedAt *string `db:"created_at" json:"createdAt"`
	RemovedAt *string `db:"removed_at" json:"removedAt"`
}

type ShippingInfo struct {
	ID             *uint64 `db:"id" json:"id"`
	Uid            *string `db:"uid" json:"uid"`
	StreetAddress1 *string `db:"street_address1" json:"streetAddress1"`
	StreetAddress2 *string `db:"street_address2" json:"streetAddress2"`
	City           *string `db:"city" json:"city"`
	State          *string `db:"state" json:"state"`
	Zip            *string `db:"zip" json:"zip"`
	CreatedAt      *string `db:"created_at" json:"createdAt"`
	DeletedAt      *string `db:"deleted_at" json:"deletedAt"`
}
