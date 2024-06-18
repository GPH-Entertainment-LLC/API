package query

var VendorItemListPageDefault = `
select
	distinct
		i.id as item_id
		, i.created_at
		, i.deleted_at
		, i.updated_at
		, i.description
		, i.name
		, i.image_url
		, i.content_main_url
		, i.content_thumb_url
		, i.content_type
		, i.active
		, i.notify
		, i.value
		, string_agg(c.category, ',') as raw_categories
		, r.rarity
		, r.ranking
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
		, count(*) over () as total_items_amount
	from
		main.items i
	left join
		main.rarity r
		on i.rarity_id = r.id
	join
		main.vendors v
		on i.vendor_id = v.uid
		and v.uid = '%v'
		and i.active = true
	left join
		main.item_categories
		on item_categories.item_id = i.id
	left join
		main.categories c
		on c.id = item_categories.category_id
	group by
		i.id, v.uid, r.id
	having
		lower(i.description) like '%%%v%%'
		or lower(i.name) like '%%%v%%'
	order by
		created_at desc
	limit %v
	offset %v;
`

var VendorItemListPageSortBy = `
select
	distinct
		i.id as item_id
		, i.created_at
		, i.deleted_at
		, i.updated_at
		, i.description
		, i.name
		, i.image_url
		, i.content_main_url
		, i.content_thumb_url
		, i.content_type
		, i.active
		, i.notify
		, i.value
		, string_agg(c.category, ',') as raw_categories
		, r.rarity
		, r.ranking
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
		, count(*) over () as total_items_amount
	from
		main.items i
	left join
		main.rarity r
		on i.rarity_id = r.id
	join
		main.vendors v
		on i.vendor_id = v.uid
		and v.uid = '%v'
		and i.active = true
	left join
		main.item_categories
		on item_categories.item_id = i.id
	left join
		main.categories c
		on c.id = item_categories.category_id
	group by
		i.id, v.uid, r.id
	having
		lower(i.description) like '%%%v%%'
		or lower(i.name) like '%%%v%%'
	order by
		%v
	limit %v
	offset %v;
`

var VendorItemListPageFilterOn = `
select
	distinct
		i.id as item_id
		, i.created_at
		, i.deleted_at
		, i.updated_at
		, i.description
		, i.name
		, i.image_url
		, i.content_main_url
		, i.content_thumb_url
		, i.content_type
		, i.active
		, i.notify
		, i.value
		, string_agg(c.category, ',') as raw_categories
		, r.rarity
		, r.ranking
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
		, count(*) over () as total_items_amount
	from
		main.items i
	left join
		main.rarity r
		on i.rarity_id = r.id
	join
		main.vendors v
		on i.vendor_id = v.uid
		and v.uid = '%v'
		and i.active = true
	left join
		main.item_categories
		on item_categories.item_id = i.id
	left join
		main.categories c
		on c.id = item_categories.category_id
	group by
		i.id, v.uid, r.id
	having
		string_agg(c.category, ',') like '%%%v%%'
		and (
			lower(i.description) like '%%%v%%'
			or lower(i.name) like '%%%v%%'
		)
	order by
		created_at desc
	limit %v
	offset %v;
`

var VendorItemListPageSortByFilterOn = `
select
	distinct
		i.id as item_id
		, i.created_at
		, i.deleted_at
		, i.updated_at
		, i.description
		, i.name
		, i.image_url
		, i.content_main_url
		, i.content_thumb_url
		, i.content_type
		, i.active
		, i.notify
		, i.value
		, string_agg(c.category, ',') as raw_categories
		, r.rarity
		, r.ranking
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
		, count(*) over () as total_items_amount
	from
		main.items i
	left join
		main.rarity r
		on i.rarity_id = r.id
	join
		main.vendors v
		on i.vendor_id = v.uid
		and v.uid = '%v'
		and i.active = true
	left join
		main.item_categories
		on item_categories.item_id = i.id
	left join
		main.categories c
		on c.id = item_categories.category_id
	group by
		i.id, v.uid, r.id
	having
		string_agg(c.category, ',') like '%%%v%%'
		and (
			lower(i.description) like '%%%v%%'
			or lower(i.name) like '%%%v%%'
		)
	order by
		%v
	limit %v
	offset %v;
`
