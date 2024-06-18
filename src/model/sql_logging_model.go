package model

type SignInLog struct {
	ID        *uint64 `db:"id" json:"id"`
	Uid       *string `db:"uid" json:"uid"`
	IpAddr    *string `db:"ip_addr" json:"ipAddr"`
	LastLogin *string `db:"last_login" json:"lastLogin"`
}

type AgeAgreementLog struct {
	ID       *uint64 `db:"id" json:"id"`
	Uid      *string `db:"uid" json:"uid"`
	IpAddr   *string `db:"ip_addr" json:"ipAddr"`
	AgreedAt *string `db:"agreed_at" json:"agreedAt"`
}

type PackPurchaseLog struct {
	ID          *uint64 `db:"id" json:"id"`
	Uid         *string `db:"uid" json:"uid"`
	IpAddr      *string `db:"ip_addr" json:"ipAddr"`
	PackId      *uint64 `db:"pack_id" json:"packId"`
	PurchasedAt *string `db:"purchased_at" json:"purchasedAt"`
}

type TokenPurchaseLog struct {
	ID            *uint64 `db:"id" json:"id"`
	Uid           *string `db:"uid" json:"uid"`
	IpAddr        *string `db:"ip_addr" json:"ipAddr"`
	TokenBundleId *uint64 `db:"token_bundle_id" json:"tokenBundleId"`
	PurchasedAt   *string `db:"purchased_at" json:"purchasedAt"`
}

type UserAccountCreationLog struct {
	ID        *uint64 `db:"id" json:"id"`
	Uid       *string `db:"uid" json:"uid"`
	IpAddr    *string `db:"ip_addr" json:"ip_addr"`
	CreatedAt *string `db:"created_at" json:"createdAt"`
}

type UserAccountDeletionLog struct {
	ID        *uint64 `db:"id" json:"id"`
	Uid       *string `db:"uid" json:"uid"`
	IpAddr    *string `db:"ip_addr" json:"ip_addr"`
	DeletedAt *string `db:"deleted_at" json:"deletedAt"`
}

type DepositLog struct {
	ID           *uint64  `db:"id" json:"id"`
	Uid          *string  `db:"uid" json:"uid"`
	IpAddr       *string  `db:"ip_addr" json:"ipAddr"`
	DollarAmount *float64 `db:"dollar_amount" json:"dollarAmount"`
	DepositedAt  *string  `db:"deposited_at" json:"depositedAt"`
}

type VendorReferralLog struct {
	ID          *uint64 `db:"id" json:"id"`
	ReferrerUid *string `db:"referrer_uid" json:"referrerUid"`
	VendorId    *string `db:"vendor_id" json:"vendorId"`
	ReferredAt  *string `db:"referred_at" json:"referredAt"`
}

type VendorApprovalLog struct {
	ID         *uint64 `db:"id" json:"id"`
	VendorId   *string `db:"vendor_id" json:"vendorId"`
	IpAddr     *string `db:"ip_addr" json:"ipAddr"`
	ApprovedAt *string `db:"approved_at" json:"approvedAt"`
}

type VendorRemovalLog struct {
	ID        *uint64 `db:"id" json:"id"`
	VendorId  *string `db:"vendor_id" json:"vendorId"`
	IpAddr    *string `db:"ip_addr" json:"ipAddr"`
	RemovedAt *string `db:"removed_at" json:"removedAt"`
}

type VendorAgreementLog struct {
	LogID       *uint64 `db:"log_id" json:"logId"`
	UID         *string `db:"uid" json:"uid"`
	IP          *string `db:"ip" json:"ip"`
	LogDatetime *string `db:"log_datetime" json:"logDatetime"`
}
