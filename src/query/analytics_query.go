package query

var TotalVendorPacksSold = `
	select
		count(*)
	from
		main.pack_facts p
	join
		main.pack_configs pc
		on p.pack_config_id = pc.id
		and pc.vendor_id = '%v'
		and p.purchased_at is not null;
`

var TotalRevenueGenerated = `
	select coalesce(
		sum(pc.token_amount) / (
			select token_amount from financial.token_currency_rate where end_date is null
		),0)::numeric(6,2) as revenue
	from
		main.pack_facts p
	join
		main.pack_configs pc
		on p.pack_config_id = pc.id
		and pc.vendor_id = '%v'
		and p.purchased_at is not null;
`

var FavoritesAmount = `
	select
		favorites_amount
	from
		main.vendors
	where
		uid = '%v';
`

var AveragePackQtyPurchased = `
	select
		coalesce(avg(qty_purchased),0)::numeric(6,2) as avg_qty_purchased
	from (
		select
			p.purchased_at
			, p.owner_id
			, count(*) as qty_purchased
		from
			main.pack_facts p
		join
			main.pack_configs pc
			on p.pack_config_id = pc.id
			and pc.vendor_id = '%v'
			and p.purchased_at is not null
		group by
			p.purchased_at, p.owner_id
	) t;
`

var MinPackQtyPurchased = `
	select
		coalesce(min(qty_purchased),0)::int as min_qty_purchased
	from (
		select
			p.purchased_at
			, p.owner_id
			, count(*) as qty_purchased
		from
			main.pack_facts p
		join
			main.pack_configs pc
			on p.pack_config_id = pc.id
			and pc.vendor_id = '%v'
			and p.purchased_at is not null
		group by
			p.purchased_at, p.owner_id
	) t;
`

var MaxPackQtyPurchased = `
	select
		coalesce(max(qty_purchased),0)::int as max_qty_purchased
	from (
		select
			p.purchased_at
			, p.owner_id
			, count(*) as qty_purchased
		from
			main.pack_facts p
		join
			main.pack_configs pc
			on p.pack_config_id = pc.id
			and pc.vendor_id = '%v'
			and p.purchased_at is not null
		group by
			p.purchased_at, p.owner_id
	) t;
`

var QtySoldPerPack = `
	select
		pc.id
		, pc.vendor_id
		, pc.created_at
		, pc.deleted_at
		, pc.updated_at
		, pc.description
		, pc.title
		, pc.token_amount
		, pc.qty
		, pc.item_qty
		, pc.current_stock
		, pc.content_main_url
		, pc.content_thumb_url
		, pc.active
		, pc.qty_sold
	from
		main.pack_configs pc
	where
		pc.vendor_id = '%v'
		and active = true
	order by
		pc.created_at desc;
`

var AvgQtySoldPerPack = `
	select
		pc.id
		, pc.vendor_id
		, pc.created_at
		, pc.deleted_at
		, pc.updated_at
		, pc.description
		, pc.title
		, pc.token_amount
		, pc.qty
		, pc.item_qty
		, pc.current_stock
		, pc.content_main_url
		, pc.content_thumb_url
		, pc.active
		, pc.qty_sold
		, avg(qty_purchased)
	from (
		select
			p.purchased_at
			, p.owner_id
			, p.pack_config_id
			, count(p.purchased_at) as qty_purchased
		from
			main.pack_configs pc
		join
			main.pack_facts p
			on pc.id = p.pack_config_id
			and p.purchased_at is not null
			and pc.vendor_id = '%v'
		group by
			p.pack_config_id, p.purchased_at, p.owner_id
	) t
	join
		main.pack_configs pc
		on t.pack_config_id = pc.id
	group by 
		pc.id
	order by 
		pc.created_at desc;
`

var TopCustomers = `
	select
		u.username
		, count(*) as packs_purchased
		, sum(o.token_amount / r.token_amount)::numeric(6,2) as amount_spent
	from
		financial.pack_orders o
	join
		financial.token_currency_rate r
		on o.token_rate_id = r.id
	join
		main.pack_facts pf
		on o.pack_id = pf.id
	join
		main.pack_configs pc
		on pf.pack_config_id = pc.id
		and pc.vendor_id = '%v'
	join
		main.users u
		on u.uid = o.uid
	group by
		u.username
	limit
		10;
`

var PackSalesByDay = `
	with date_series as (
		select generate_series(
			date_trunc('month', '%v'::date),
			(date_trunc('month', '%v'::date) + INTERVAL '1 month' - INTERVAL '1 day'),
			'1 day'::interval
			) AS order_date
		)

	select
		date_series.order_date::date as granularity
		, count(o.ordered_at) as qty_sold
		, sum(o.token_amount / r.token_amount)::numeric(6,2) as total_sales
	from
 		financial.pack_orders o
	join
		financial.token_currency_rate r
		on o.token_rate_id = r.id
	join
		main.pack_facts p
		on o.pack_id = p.id
	join
		main.pack_configs pc
		on p.pack_config_id = pc.id
		and pc.vendor_id = '%v'
	right join
		date_series
		on date_series.order_date = o.ordered_at::date
	group by
 		date_series.order_date
	order by
 		date_series.order_date asc;
`

var PackSalesByMonth = `
	with month_series as (
		select generate_series(
		1,
		12,
		1
		) AS order_month
	)

	select
		month_series.order_month as granularity
		, count(o.ordered_at) qty_sold
		, sum(o.token_amount / r.token_amount)::numeric(6,2) as total_sales
	from
		financial.pack_orders o
	join
		financial.token_currency_rate r
		on o.token_rate_id = r.id
		and extract(year from o.ordered_at) = %v
	join
		main.pack_facts p
		on o.pack_id = p.id
	join
		main.pack_configs pc
		on p.pack_config_id = pc.id
		and pc.vendor_id = '%v'
	right join
		month_series
		on month_series.order_month = extract(month from o.ordered_at)
	group by
		month_series.order_month
	order by
		month_series.order_month asc;
`

var PackSalesByYear = `
	with year_series as (
		select generate_series(
			extract(year from current_date) - 4,
			extract(year from current_date),
			1
  		) AS order_year
	)

	select
		year_series.order_year as granularity
		, count(o.ordered_at) as qty_sold
		, sum(o.token_amount / r.token_amount)::numeric(6,2) as total_sales
	from
 		financial.pack_orders o
	join
		financial.token_currency_rate r
		on o.token_rate_id = r.id
	join
		main.pack_facts p
		on o.pack_id = p.id
	join
		main.pack_configs pc
		on p.pack_config_id = pc.id
		and pc.vendor_id = '%v'
	right join
 		year_series
 		on year_series.order_year = extract(year from o.ordered_at)
	group by
 		year_series.order_year
	order by
 		year_series.order_year asc;
`

var PackQtySold = `
	select
		title as pack_title
		, qty_sold
	from
		main.pack_configs
	where
		vendor_id = '%v'
		and active = true;
`
