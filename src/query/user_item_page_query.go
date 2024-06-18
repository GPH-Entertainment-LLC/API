package query

var UserItemPageDefault = `
	select
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
		, v.display_name as vendor_display_name
		, v.first_name as vendor_first_name
		, v.last_name as vendor_last_name
		, v.username as vendor_username
		, v.image_url as vendor_image_url
		, v.bio as vendor_bio
		, v.favorites_amount as vendor_favorites_amount
		, v.active as vendor_active
		, v.content_main_url as vendor_main_url
		, v.content_thumb_url as vendor_thumb_url
		, v.banner_url as vendor_banner_url
		, ui.id as user_item_id
		, ui.uid as owner_id
		, ui.acquired_at
		, count(*) over () as total_items_amount
	from
		main.items i
	left join
		main.rarity r
		on i.rarity_id = r.id
	join
		main.vendors v
		on i.vendor_id = v.uid
	join
		main.user_items ui
		on i.id = ui.item_id
		and ui.uid = '%v'
		and ui.removed_at is null
	left join
		main.item_categories
		on item_categories.item_id = i.id
	left join
		main.categories c
		on c.id = item_categories.category_id
	group by
		i.id, v.uid, r.id, ui.id
	having
		lower(v.username) like '%%%v%%'
		or lower(v.first_name) like '%%%v%%'
		or lower(v.last_name) like '%%%v%%'
		or lower(v.display_name) like '%%%v%%'
		or lower(i.name) like '%%%v%%'
	order by
		acquired_at desc
	limit %v
	offset %v;
`

var UserItemPageSortBy = `
	select
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
		, v.display_name as vendor_display_name
		, v.first_name as vendor_first_name
		, v.last_name as vendor_last_name
		, v.username as vendor_username
		, v.image_url as vendor_image_url
		, v.bio as vendor_bio
		, v.favorites_amount as vendor_favorites_amount
		, v.active as vendor_active
		, v.content_main_url as vendor_main_url
		, v.content_thumb_url as vendor_thumb_url
		, v.banner_url as vendor_banner_url
		, ui.id as user_item_id
		, ui.uid as owner_id
		, ui.acquired_at
		, count(*) over () as total_items_amount
	from
		main.items i
	left join
		main.rarity r
		on i.rarity_id = r.id
	join
		main.vendors v
		on i.vendor_id = v.uid
	join
		main.user_items ui
		on i.id = ui.item_id
		and ui.uid = '%v'
		and ui.removed_at is null
	left join
		main.item_categories
		on item_categories.item_id = i.id
	left join
		main.categories c
		on c.id = item_categories.category_id
	group by
		i.id, v.uid, r.id, ui.id
	having
		lower(v.username) like '%%%v%%'
		or lower(v.first_name) like '%%%v%%'
		or lower(v.last_name) like '%%%v%%'
		or lower(v.display_name) like '%%%v%%'
		or lower(i.name) like '%%%v%%'
	order by
		%v
	limit %v
	offset %v;
`

var UserItemPageFilterOn = `	
	select
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
		, v.display_name as vendor_display_name
		, v.first_name as vendor_first_name
		, v.last_name as vendor_last_name
		, v.username as vendor_username
		, v.image_url as vendor_image_url
		, v.bio as vendor_bio
		, v.favorites_amount as vendor_favorites_amount
		, v.active as vendor_active
		, v.content_main_url as vendor_main_url
		, v.content_thumb_url as vendor_thumb_url
		, v.banner_url as vendor_banner_url
		, ui.id as user_item_id
		, ui.uid as owner_id
		, ui.acquired_at
		, count(*) over () as total_items_amount
	from
		main.items i
	left join
		main.rarity r
		on i.rarity_id = r.id
	join
		main.vendors v
		on i.vendor_id = v.uid
	join
		main.user_items ui
		on i.id = ui.item_id
		and ui.uid = '%v'
		and ui.removed_at is null
	left join
		main.item_categories
		on item_categories.item_id = i.id
	left join
		main.categories c
		on c.id = item_categories.category_id
	group by
		i.id, v.uid, r.id, ui.id
	having
		string_agg(c.category, ',') like '%%%v%%'
		and (
			lower(v.username) like '%%%v%%'
			or lower(v.first_name) like '%%%v%%'
			or lower(v.last_name) like '%%%v%%'
			or lower(v.display_name) like '%%%v%%'
			or lower(i.name) like '%%%v%%')
	order by
		acquired_at desc
	limit %v
	offset %v;	
`

var UserItemPageSortByFilterOn = `
	select
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
		, v.display_name as vendor_display_name
		, v.first_name as vendor_first_name
		, v.last_name as vendor_last_name
		, v.username as vendor_username
		, v.image_url as vendor_image_url
		, v.bio as vendor_bio
		, v.favorites_amount as vendor_favorites_amount
		, v.active as vendor_active
		, v.content_main_url as vendor_main_url
		, v.content_thumb_url as vendor_thumb_url
		, v.banner_url as vendor_banner_url
		, ui.id as user_item_id
		, ui.uid as owner_id
		, ui.acquired_at
		, count(*) over () as total_items_amount
	from
		main.items i
	left join
		main.rarity r
		on i.rarity_id = r.id
	join
		main.vendors v
		on i.vendor_id = v.uid
	join
		main.user_items ui
		on i.id = ui.item_id
		and ui.uid = '%v'
		and ui.removed_at is null
	left join
		main.item_categories
		on item_categories.item_id = i.id
	left join
		main.categories c
		on c.id = item_categories.category_id
	group by
		i.id, v.uid, r.id, ui.id
	having
		string_agg(c.category, ',') like '%%%v%%'
		and (
			lower(v.username) like '%%%v%%'
			or lower(v.first_name) like '%%%v%%'
			or lower(v.last_name) like '%%%v%%'
			or lower(v.display_name) like '%%%v%%'
			or lower(i.name) like '%%%v%%')
	order by
		%v
	limit %v
	offset %v;
`
