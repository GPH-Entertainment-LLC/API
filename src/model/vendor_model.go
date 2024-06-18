package model

type Vendor struct {
	UID               *string  `db:"uid" json:"uid" `
	Email             *string  `db:"email" json:"email"`
	CreatedAt         *string  `db:"created_at" json:"createdAt"`
	UpdatedAt         *string  `db:"updated_at" json:"updatedAt"`
	DeletedAt         *string  `db:"deleted_at" json:"deletedAt"`
	ImageUrl          *string  `db:"image_url" json:"imageUrl"`
	FirstName         *string  `db:"first_name" json:"firstName"`
	LastName          *string  `db:"last_name" json:"lastName"`
	DisplayName       *string  `db:"display_name" json:"displayName"`
	Username          *string  `db:"username" json:"username"`
	Birthday          *string  `db:"birthday" json:"birthday"`
	Bio               *string  `db:"bio" json:"bio"`
	IsVendor          *bool    `db:"is_vendor" json:"isVendor"`
	Active            *bool    `db:"active" json:"active"`
	PackAmount        *uint64  `db:"pack_amount" json:"packAmount"`
	FavoritesAmount   *uint64  `db:"favorites_amount" json:"favoritesAmount"`
	RawCategories     *string  `db:"categories" json:"rawCategories"`
	Categories        []string `json:"categories"`
	ContentMainUrl    *string  `db:"content_main_url" json:"contentMainUrl"`
	ContentThumbUrl   *string  `db:"content_thumb_url" json:"contentThumbUrl"`
	BannerUrl         *string  `db:"banner_url" json:"bannerUrl"`
	ProfileImgVersion *int     `db:"profile_img_version" json:"profileImgVersion"`
	BannerImgVersion  *int     `db:"banner_img_version" json:"bannerImgVersion"`
}
