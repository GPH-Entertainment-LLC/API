package query

var VendorPackListPageDefault = `
	select
	distinct
		pc.id as pack_config_id
		, pc.image_url
		, pc.item_qty
		, pc.description
		, pc.title
		, pc.active
		, pc.created_at
		, pc.updated_at
		, pc.deleted_at
		, pc.release_at
		, pc.token_amount
		, pc.qty
		, string_agg(c.category, ',') as raw_categories
		, pc.current_stock
		, pc.qty_sold
		, pc.content_main_url
		, pc.content_thumb_url
		, v.uid as vendor_id
		, v.first_name as vendor_first_name
		, v.last_name as vendor_last_name
		, v.username as vendor_username
		, v.display_name as vendor_display_name
		, v.image_url as vendor_image_url
		, v.favorites_amount as vendor_favorites_amount
		, v.bio as vendor_bio
		, v.active as vendor_active
		, v.content_main_url as vendor_main_url
		, v.content_thumb_url as vendor_thumb_url
		, v.banner_url as vendor_banner_url
		, count(*) over () as total_packs_amount
	from
		main.pack_configs pc
	join
		main.vendors v
		on v.uid = pc.vendor_id
		and pc.deleted_at is null
		and v.uid = '%v'
		and v.active = true
	left join
		main.pack_categories
		on pack_categories.pack_config_id = pc.id
	left join
		main.categories c
		on c.id = pack_categories.category_id
	group by
		pc.id, v.uid
	having
		lower(pc.title) like '%%%v%%'
		or lower(pc.description) like '%%%v%%'
	order by
		created_at desc
	limit %v
	offset %v;
`

var VendorPackListPageSortBy = `
	select
	distinct
		pc.id as pack_config_id
		, pc.image_url
		, pc.item_qty
		, pc.description
		, pc.title
		, pc.active
		, pc.created_at
		, pc.updated_at
		, pc.deleted_at
		, pc.release_at
		, pc.token_amount
		, pc.qty
		, string_agg(c.category, ',') as raw_categories
		, pc.current_stock
		, pc.qty_sold
		, pc.content_main_url
		, pc.content_thumb_url
		, v.uid as vendor_id
		, v.first_name as vendor_first_name
		, v.last_name as vendor_last_name
		, v.username as vendor_username
		, v.display_name as vendor_display_name
		, v.image_url as vendor_image_url
		, v.favorites_amount as vendor_favorites_amount
		, v.bio as vendor_bio
		, v.active as vendor_active
		, v.content_main_url as vendor_main_url
		, v.content_thumb_url as vendor_thumb_url
		, v.banner_url as vendor_banner_url
		, count(*) over () as total_packs_amount
	from
		main.pack_configs pc
	join
		main.vendors v
		on v.uid = pc.vendor_id
		and pc.deleted_at is null
		and v.uid = '%v'
		and v.active = true
	left join
		main.pack_categories
		on pack_categories.pack_config_id = pc.id
	left join
		main.categories c
		on c.id = pack_categories.category_id
	group by
		pc.id, v.uid
	having
		lower(pc.title) like '%%%v%%'
		or lower(pc.description) like '%%%v%%'
	order by
		%v
	limit %v
	offset %v;
`

var VendorPackListPageFilterOn = `
	select
	distinct
		pc.id as pack_config_id
		, pc.image_url
		, pc.item_qty
		, pc.description
		, pc.title
		, pc.active
		, pc.created_at
		, pc.updated_at
		, pc.deleted_at
		, pc.release_at
		, pc.token_amount
		, pc.qty
		, string_agg(c.category, ',') as raw_categories
		, pc.current_stock
		, pc.qty_sold
		, pc.content_main_url
		, pc.content_thumb_url
		, v.uid as vendor_id
		, v.first_name as vendor_first_name
		, v.last_name as vendor_last_name
		, v.username as vendor_username
		, v.display_name as vendor_display_name
		, v.image_url as vendor_image_url
		, v.favorites_amount as vendor_favorites_amount
		, v.bio as vendor_bio
		, v.active as vendor_active
		, v.content_main_url as vendor_main_url
		, v.content_thumb_url as vendor_thumb_url
		, v.banner_url as vendor_banner_url
		, count(*) over () as total_packs_amount
	from
		main.pack_configs pc
	join
		main.vendors v
		on v.uid = pc.vendor_id
		and pc.deleted_at is null
		and v.uid = '%v'
		and v.active = true
	left join
		main.pack_categories
		on pack_categories.pack_config_id = pc.id
	left join
		main.categories c
		on c.id = pack_categories.category_id
	group by
		pc.id, v.uid
	having
		string_agg(c.category, ',') like '%%%v%%'
		and (
			lower(pc.title) like '%%%v%%'
			or lower(pc.description) like '%%%v%%'
		)
	order by
		created_at desc
	limit %v
	offset %v;
`

var VendorPackListPageSortByFilterOn = `
	select
	distinct
		pc.id as pack_config_id
		, pc.image_url
		, pc.item_qty
		, pc.description
		, pc.title
		, pc.active
		, pc.created_at
		, pc.updated_at
		, pc.deleted_at
		, pc.release_at
		, pc.token_amount
		, pc.qty
		, string_agg(c.category, ',') as raw_categories
		, pc.current_stock
		, pc.qty_sold
		, pc.content_main_url
		, pc.content_thumb_url
		, v.uid as vendor_id
		, v.first_name as vendor_first_name
		, v.last_name as vendor_last_name
		, v.username as vendor_username
		, v.display_name as vendor_display_name
		, v.image_url as vendor_image_url
		, v.favorites_amount as vendor_favorites_amount
		, v.bio as vendor_bio
		, v.active as vendor_active
		, v.content_main_url as vendor_main_url
		, v.content_thumb_url as vendor_thumb_url
		, v.banner_url as vendor_banner_url
		, count(*) over () as total_packs_amount
	from
		main.pack_configs pc
	join
		main.vendors v
		on v.uid = pc.vendor_id
		and pc.deleted_at is null
		and v.uid = '%v'
		and v.active = true
	left join
		main.pack_categories
		on pack_categories.pack_config_id = pc.id
	left join
		main.categories c
		on c.id = pack_categories.category_id
	group by
		pc.id, v.uid
	having
		string_agg(c.category, ',') like '%%%v%%'
		and (
			lower(pc.title) like '%%%v%%'
			or lower(pc.description) like '%%%v%%'
		)
	order by
		%v
	limit %v
	offset %v;
`
