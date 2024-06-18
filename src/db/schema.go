package db

// DB SCHEMAS
const (
	SCHEMA_USERS                      = "main.users"
	SCHEMA_VENDORS                    = "main.vendors"
	SCHEMA_ITEMS                      = "main.items"
	SCHEMA_VENDOR_CATEGORIES          = "main.vendor_categories"
	SCHEMA_CATEGORIES                 = "main.categories"
	SCHEMA_SORT_MAPPINGS              = "main.sort_mappings"
	SCHEMA_USER_ITEMS                 = "main.user_items"
	SCHEMA_ITEM_CATEGORIES            = "main.item_categories"
	SCHEMA_PACK_FACTS                 = "main.pack_facts"
	SCHEMA_PACK_CONFIGS               = "main.pack_configs"
	SCHEMA_PACK_ITEM_CONFIGS          = "main.pack_item_configs"
	SCHEMA_PACK_ITEM_FACTS            = "main.pack_item_facts"
	SCHEMA_PACK_CATEGORIES            = "main.pack_categories"
	SCHEMA_TOKEN_BUNDLE               = "financial.token_bundle"
	SCHEMA_TOKEN_BALANCE              = "financial.token_balance"
	SCHEMA_TOKEN_ORDERS               = "financial.token_orders"
	SCHEMA_PACK_ORDERS                = "financial.pack_orders"
	SCHEMA_NEW_SALES_TRANSACTIONS     = "financial.new_sales_transactions"
	SCHEMA_TRANSACTIONS               = "financial.transactions"
	SCHEMA_ITEM_WITHDRAWALS           = "financial.item_withdrawals"
	SCHEMA_SIGN_INS                   = "logging.sign_ins"
	SCHEMA_AGE_AGREEMENTS             = "logging.age_agreements"
	SCHEMA_USER_ACCOUNT_CREATION_LOGS = "logging.user_account_creation_logs"
	SCHEMA_USER_ACCOUNT_DELETION_LOGS = "logging.user_account_deletion_logs"
	SCHEMA_PACK_PURCHASES             = "logging.pack_purchases"
	SCHEMA_TOKEN_PURCHASES            = "logging.token_purchases"
	SCHEMA_DEPOSITS                   = "logging.deposits"
	SCHEMA_VENDOR_APPROVAL_LOGS       = "logging.vendor_approval_logs"
	SCHEMA_VENDOR_REMOVAL_LOGS        = "logging.vendor_removal_logs"
	SCHEMA_VENDOR_REFERRAL_LOGS       = "logging.vendor_referral_log"
	SCHEMA_FULFILLABLE_PULLS          = "logging.fulfillable_pulls"
	SCHEMA_NON_FULFILLABLE_PULLS      = "logging.non_fulfillable_pulls"
	SCHEMA_VENDOR_AGREEMENT_LOG       = "logging.vendor_agreement_log"
	SCHEMA_FAVORITES                  = "main.favorites"
	SCHEMA_VENDOR_APPLICATIONS        = "main.vendor_applications"
	SCHEMA_REFERRAL_CODES             = "main.referral_codes"
	SCHEMA_REFERRALS                  = "main.referrals"
	SCHEMA_REPORT_OPTS                = "main.report_opts"
	SCHEMA_REPORTS                    = "main.reports"
	SCHEMA_FAQS                       = "main.faqs"
)

// CACHE KEYS
const (
	KEY_USERNAME              = "user_username_"
	KEY_USER                  = "user_"
	KEY_PRIVATE_USER          = "private_user_"
	KEY_VENDOR                = "vendor_"
	KEY_PACK                  = "pack_"
	KEY_PACK_CONFIG           = "pack_config_"
	KEY_PACK_ITEMS            = "pack_items_"
	KEY_PACK_ITEM_CONFIGS     = "pack_item_configs_"
	KEY_PACK_VENDOR_CONFIGS   = "pack_vendor_configs_"
	KEY_ITEM                  = "item_"
	KEY_USER_ITEM             = "user_item_"
	KEY_ITEM_WITHDRAWALS      = "item_withdrawals_"
	KEY_USER_ITEMS            = "user_items_"
	KEY_USER_ITEM_AMOUNT      = "user_item_amount_"
	KEY_USER_PACK_AMOUNT      = "user_pack_amount_"
	KEY_CATEGORY              = "category_"
	KEY_ALL_CATEGORY          = "categories"
	KEY_ALL_CATEGORY_LITERALS = "category_literals"
	KEY_TOKEN_BUNDLE          = "token_bundle_"
	KEY_TOKEN_BUNDLES         = "token_bundles_"
	KEY_TOKEN_BALANCE         = "token_balance_"
	KEY_FAVORITES             = "_favorite_"
	KEY_VENDOR_CATEGORIES     = "vendor_categories_"
	KEY_TOTAL_PACKS_SOLD      = "packs_sold_"
	KEY_TOTAL_REVENUE         = "total_revenue_"
	KEY_AVG_PURCHASED_QTY     = "avg_purchased_qty_"
	KEY_MIN_PURCHASED_QTY     = "min_purchased_qty_"
	KEY_MAX_PURCHASED_QTY     = "max_purchased_qty_"
	KEY_FAVORITE_AMOUNT       = "favorite_amount_"
	KEY_ACTIVE_TOKEN_RATE     = "active_token_rate"
	KEY_TOP_CUSTOMERS         = "top_customers_"
	KEY_PACK_SALES            = "_pack_sales_"
	KEY_PACK_QTY              = "pack_qty_sold_"
	KEY_PEM                   = "pem_key"
	KEY_KEY_PAIR_ID           = "key_pair_id"
	// KEY_CREATOR_PENDING_APPLICATION   = "creator_pending_application_"
	// KEY_CREATOR_APPROVED_APPLICATION  = "creator_approved_application_"
	// KEY_CREATOR_REJECTED_APPLICATION  = "creator_approved_application_"
	// KEY_ACTIVE_CREATOR_REFERRAL_CODES = "active_creator_referral_codes_"
	KEY_PENDING_APPLICATION        = "pending_application_"
	KEY_APPROVED_APPLICATION       = "approved_application_"
	KEY_REJECTED_APPLICATION       = "rejected_application_"
	KEY_ACTIVE_USER_REFERRAL_CODES = "active_user_referral_codes_"
	KEY_ACTIVE_REFERRAL_CODES      = "active_referral_codes_"
	KEY_REPORT_OPTS                = "report_opts"
)

// LOG MSG HEADERS
const (
	LOG_USER_CREATE             = "client_logs_new_user_log"
	LOG_USER_DELETE             = "client_logs_user_delete_log"
	LOG_APPLICATION             = "client_logs_application_log"
	LOG_APPLICATION_STATUS      = "client_logs_application_status_log"
	LOG_VENDOR_REMOVAL          = "client_logs_vendor_removal_log"
	LOG_PACK_OPEN               = "client_logs_pack_opens_log"
	LOG_ITEM_PULLS              = "client_logs_item_pulls_log"
	LOG_ITEM_NOTIFY_PULLS       = "client_logs_item_notify_pulls_log"
	LOG_ITEM_DATA_QUALITY_ERROR = "client_logs_item_data_quality_errors_log"
	LOG_SIGN_IN                 = "client_logs_sign_in_log"
	LOG_ADMIN_LOG               = "admin_login"
	LOG_REFERRAL                = "client_logs_referral_log"
	LOG_REPORT                  = "client_logs_report_log"
)
