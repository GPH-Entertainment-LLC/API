package query

var UserTransactionHistoryPageQuery = `	
	select
		t.uid
		, t.token_bundle_id
		, t.subscription_id
		, t.transaction_id
		, t.tran_datetime
		, t.client_accnum
		, t.client_subacc
		, t.billed_initial_price
		, t.initial_period
		, t.billed_recurring_price
		, t.recurring_period
		, t.rebills
		, t.currency_code
		, tb.dollar_amount
		, tb.token_amount
		, tb.bundle_image_url
		, count(*) over () as transaction_count
	from
		financial.transactions t
	join
		financial.token_bundle tb
		on t.token_bundle_id = tb.id
	where
		uid = '%v'
	order by
		t.tran_datetime desc
	limit
		%v
	offset
		%v;
`
