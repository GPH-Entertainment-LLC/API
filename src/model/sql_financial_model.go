package model

type TokenBundle struct {
	ID             *uint64  `db:"id" json:"id"`
	DollarAmount   *float64 `db:"dollar_amount" json:"dollarAmount"`
	TokenAmount    *float64 `db:"token_amount" json:"tokenAmount"`
	CreatedAt      *string  `db:"created_at" json:"createdAt"`
	DeletedAt      *string  `db:"deleted_at" json:"deletedAt"`
	Active         *bool    `db:"active" json:"active"`
	BundleImageUrl *string  `db:"bundle_image_url" json:"bundleImageUrl"`
}

type TokenBalance struct {
	ID        *uint64  `db:"id" json:"id"`
	UID       *string  `db:"uid" json:"uid"`
	Balance   *float64 `db:"balance" json:"balance"`
	UpdatedAt *string  `db:"updated_at" json:"updatedAt"`
}

type PackOrder struct {
	ID          *uint64  `db:"id" json:"id"`
	PackId      *uint64  `db:"pack_id" json:"packId"`
	Uid         *string  `db:"uid" json:"uid"`
	TokenAmount *float64 `db:"token_amount" json:"tokenAmount"`
	TokenRateId *uint64  `db:"token_rate_id" json:"tokenRateId"`
	OrderedAt   *string  `db:"ordered_at" json:"orderedAt"`
}

type TokenOrder struct {
	ID            *uint64  `db:"id" json:"id"`
	Uid           *string  `db:"uid" json:"uid"`
	TransactionId *string  `db:"transaction_id" json:"transactionId"`
	TokenBundleId *uint64  `db:"token_bundle_id" json:"tokenBundleId"`
	PriceUsd      *float64 `db:"price_usd" json:"priceUsd"`
	TokenRateId   *uint64  `db:"token_rate_id" json:"tokenRateId"`
	OrderedAt     *string  `db:"ordered_at" json:"orderedAt"`
}

type TokenCurrencyRate struct {
	ID           *uint64  `db:"id" json:"id"`
	TokenAmount  *float64 `db:"token_amount" json:"tokenAmount"`
	DollarAmount *float64 `db:"dollar_amount" json:"dollarAmount"`
	StartDate    *string  `db:"start_date" json:"startDate"`
	EndDate      *string  `db:"end_date" json:"endDate"`
}

type ItemWithdrawal struct {
	ID          *uint64 `db:"id" json:"id"`
	UserItemId  *uint64 `db:"user_item_id" json:"userItemId"`
	WithdrawnAt *string `db:"withdrawn_at" json:"withdrawnAt"`
	FulfilledAt *string `db:"fulfilled_at" json:"fulfilledAt"`
}

type NewSalesTransaction struct {
	Uid                            *string  `db:"uid" json:"uid"`
	TokenBundleId                  *uint64  `db:"token_bundle_id" json:"X-tokenBundleId"`
	SubscriptionId                 *int     `db:"subscription_id" json:"subscriptionId"`
	TransactionId                  *string  `db:"transaction_id" json:"transactionId"`
	CurrencyCode                   *int     `db:"currency_code" json:"X-currencyCode"`
	FormDigest                     *string  `db:"form_digest" json:"X-formDigest"`
	ClientAccNum                   *string  `db:"client_accnum" json:"clientAccnum"`
	ClientSubAcc                   *string  `db:"client_subacc" json:"clientSubacc"`
	TranDatetime                   *string  `db:"tran_datetime" json:"timestamp"`
	FirstName                      *string  `db:"first_name" json:"firstName"`
	LastName                       *string  `db:"last_name" json:"lastName"`
	Address1                       *string  `db:"address1" json:"address1"`
	City                           *string  `db:"city" json:"city"`
	State                          *string  `db:"state" json:"state"`
	Country                        *string  `db:"country" json:"country"`
	PostalCode                     *string  `db:"postal_code" json:"postalCode"`
	Email                          *string  `db:"email" json:"email"`
	PhoneNumber                    *string  `db:"phone_number" json:"phoneNumber"`
	IpAddress                      *string  `db:"ip_address" json:"ipAddress"`
	ReservationId                  *string  `db:"reservation_id" json:"reservationId"`
	Username                       *string  `db:"username" json:"username"`
	Password                       *string  `db:"password" json:"password"`
	FormName                       *string  `db:"form_name" json:"formName"`
	FlexId                         *string  `db:"flex_id" json:"flexId"`
	ProductDesc                    *string  `db:"product_desc" json:"productDescription"`
	PriceDesc                      *string  `db:"price_desc" json:"priceDescription"`
	RecurringPriceDesc             *string  `db:"recurring_price_desc" json:"recurringPriceDescription"`
	BilledInitialPrice             *string  `db:"billed_initial_price" json:"billedInitialPrice"`
	BilledRecurringPrice           *string  `db:"billed_recurring_price" json:"billedRecurringPrice"`
	BilledCurrencyCode             *int     `db:"billed_currency_code" json:"billedCurrencyCode"`
	SubscriptionInitialPrice       *float64 `db:"subscription_initial_price" json:"subscriptionInitialPrice"`
	SubscriptionRecurringPrice     *float64 `db:"subscription_recurring_price" json:"subscriptionRecurringPrice"`
	SubscriptionCurrencyCode       *int     `db:"subscription_currency_code" json:"subscriptionCurrencyCode"`
	AccountingInitialPrice         *float64 `db:"accounting_initial_price" json:"accountingInitialPrice"`
	AccountRecurringPrice          *float64 `db:"account_recurring_price" json:"accountRecurringPrice"`
	AccountingCurrencyCode         *int     `db:"accounting_currency_code" json:"accountingCurrencyCode"`
	InitialPeriod                  *int     `db:"initial_period" json:"initialPeriod"`
	RecurringPeriod                *int     `db:"recurring_period" json:"recurringPeriod"`
	Rebills                        *int     `db:"rebills" json:"rebills"`
	NextRenewalDate                *string  `db:"next_renewal_date" json:"nextRenewalDate"`
	SubscriptionTypeId             *int     `db:"subscription_type_id" json:"subscriptionTypeId"`
	DynamicPricingValidationDigest *string  `db:"dynamic_pricing_validation_digest" json:"dynamicPricingValidationDigest"`
	PaymentType                    *string  `db:"payment_type" json:"paymentType"`
	CardType                       *string  `db:"card_type" json:"cardType"`
	Bin                            *int     `db:"bin" json:"bin"`
	PrePaid                        *bool    `db:"pre_paid" json:"prePaid"`
	Last4                          *int     `db:"last4" json:"last4"`
	ExpDate                        *int     `db:"exp_date" json:"expDate"`
	AvsResp                        *string  `db:"avs_resp" json:"avsResponse"`
	Cvv2Resp                       *string  `db:"cvv2_resp" json:"cvv2Response"`
	AffiliateSystem                *string  `db:"affiliate_system" json:"affiliateSystem"`
	ReferringUrl                   *string  `db:"referring_url" json:"referringUrl"`
	PaymentAcc                     *string  `db:"payment_acc" json:"paymentAccount"`
	ThreeDSecure                   *string  `db:"three_d_secure" json:"threeDSecure"`
}

type Transaction struct {
	Uid                    string  `db:"uid" json:"uid"`
	TokenBundleId          int     `db:"token_bundle_id" json:"tokenBundleId"`
	TokenAmount            float64 `json:"tokenAmount"`
	TransactionId          string  `db:"transaction_id" json:"transactionId"`
	SubscriptionId         string  `db:"subscription_id" json:"subscriptionId"`
	TranDatetime           string  `db:"tran_datetime" json:"tranDateTime"`
	ClientAccNum           string  `db:"client_accnum" json:"clientAccnum"`
	ClientSubAcc           string  `db:"client_subacc" json:"clientSubacc"`
	Username               string  `db:"username" json:"username"`
	Password               string  `db:"password" json:"password"`
	InitialPrice           string  `db:"billed_initial_price" json:"initialPrice"`
	InitialPeriod          string  `db:"initial_period" json:"initialPeriod"`
	RecurringPrice         string  `db:"billed_recurring_price" json:"recurringPrice"`
	RecurringPeriod        string  `db:"recurring_period" json:"recurringPeriod"`
	Rebills                string  `db:"rebills" json:"rebills"`
	CurrencyCode           string  `db:"currency_code" json:"currencyCode"`
	PaymentProcessingFeeId int     `db:"payment_processing_fee_id" json:"paymentProcessingFeeId"`
}

type UserTransactionInfo struct {
	Uid             *string `db:"uid" json:"uid"`
	TokenBundleId   *int    `db:"token_bundle_id" json:"tokenBundleId"`
	TransactionId   *string `db:"transaction_id" json:"transactionId"`
	SubscriptionId  *string `db:"subscription_id" json:"subscriptionId"`
	ClientAccNum    *string `db:"client_accnum" json:"clientAccnum"`
	ClientSubAcc    *string `db:"client_subacc" json:"clientSubacc"`
	Username        *string `db:"username" json:"username"`
	Password        *string `db:"password" json:"password"`
	Last4           *int    `db:"last4" json:"last4"`
	PaymentType     *string `db:"payment_type" json:"paymentType"`
	CardType        *string `db:"card_type" json:"cardType"`
	InitialPrice    *string `db:"billed_initial_price" json:"initialPrice"`
	InitialPeriod   *int    `db:"initial_period" json:"initialPeriod"`
	RecurringPrice  *string `db:"billed_recurring_price" json:"recurringPrice"`
	RecurringPeriod *int    `db:"recurring_period" json:"recurringPeriod"`
	Rebills         *int    `db:"rebills" json:"rebills"`
	CurrencyCode    *int    `db:"currency_code" json:"currencyCode"`
}

type CreatorEarningsPeriod struct {
	Uid            *string  `db:"uid" json:"uid"`
	Email          *string  `db:"email" json:"email"`
	StartingPeriod *string  `db:"starting_period" json:"startingPeriod"`
	EndingPeriod   *string  `db:"ending_period" json:"endingPeriod"`
	PayoutDate     *string  `db:"payout_date" json:"payoutDate"`
	EarningType    *string  `db:"earning_type" json:"earningType"`
	Earnings       *float64 `db:"earnings" json:"earnings"`
	CurrentPeriod  *string  `db:"current_period" json:"currentPeriod"`
	PayoutStatus   *string  `db:"payout_status" json:"payoutStatus"`
}

type ReferralEarningsPeriod struct {
	Uid            *string  `db:"uid" json:"uid"`
	Email          *string  `db:"email" json:"email"`
	StartingPeriod *string  `db:"starting_period" json:"startingPeriod"`
	EndingPeriod   *string  `db:"ending_period" json:"endingPeriod"`
	PayoutDate     *string  `db:"payout_date" json:"payoutDate"`
	EarningType    *string  `db:"earning_type" json:"earningType"`
	Earnings       *float64 `db:"earnings" json:"earnings"`
	CurrentPeriod  bool     `db:"current_period" json:"currentPeriod"`
	PayoutStatus   *string  `db:"payout_status" json:"payoutStatus"`
}

type AllEarningsPeriod struct {
	Uid            *string  `db:"uid" json:"uid"`
	Email          *string  `db:"email" json:"email"`
	StartingPeriod *string  `db:"starting_period" json:"startingPeriod"`
	EndingPeriod   *string  `db:"ending_period" json:"endingPeriod"`
	PayoutDate     *string  `db:"payout_date" json:"payoutDate"`
	EarningType    *string  `db:"earning_type" json:"earningType"`
	TotalEarnings  *float64 `db:"earnings" json:"earnings"`
	CurrentPeriod  bool     `db:"current_period" json:"currentPeriod"`
	PayoutStatus   *string  `db:"payout_status" json:"payoutStatus"`
}
