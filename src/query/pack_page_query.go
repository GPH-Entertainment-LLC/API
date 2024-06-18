package query

var PackPageDefault = `
	select
	distinct
		pc.id as pack_config_id
		, pc.image_url
		, pc.item_qty
		, pc.description
		, pc.title
		, pc.vendor_id
		, pc.created_at
		, pc.token_amount
		, pc.qty
		, pc.content_main_url
		, pc.content_thumb_url
		, v.display_name as vendor_display_name
		, v.first_name as vendor_first_name
		, v.last_name as vendor_last_name
		, v.username as vendor_username
		, v.favorites_amount as vendor_favorites_amount
		, v.content_main_url as vendor_main_url
		, v.content_thumb_url as vendor_thumb_url
		, v.banner_url as vendor_banner_url
		, string_agg(c.category, ',') as raw_categories
		, count(*) over () as pack_amount
	from
		main.pack_configs pc
	join
		main.vendors v
		on v.uid = pc.vendor_id
		and v.active = true
	left join
		main.pack_categories
		on pack_categories.pack_config_id = pc.id
	left join
		main.categories c
		on c.id = pack_categories.category_id
	group by
		pc.id, v.uid
	order by
		created_at desc
	limit %v
	offset %v;
`

var PackPageSortBy = `
	select
	distinct
		pc.id as pack_config_id
		, pc.image_url
		, pc.item_qty
		, pc.description
		, pc.title
		, pc.vendor_id
		, pc.created_at
		, pc.token_amount
		, pc.qty
		, pc.content_main_url
		, pc.content_thumb_url
		, v.display_name as vendor_display_name
		, v.first_name as vendor_first_name
		, v.last_name as vendor_last_name
		, v.username as vendor_username
		, v.favorites_amount as vendor_favorites_amount
		, v.content_main_url as vendor_main_url
		, v.content_thumb_url as vendor_thumb_url
		, v.banner_url as vendor_banner_url
		, string_agg(c.category, ',') as raw_categories
		, count(*) over () as pack_amount
	from
		main.pack_configs pc
	join
		main.vendors v
		on v.uid = pc.vendor_id
		and v.active = true
	left join
		main.pack_categories
		on pack_categories.pack_config_id = pc.id
	left join
		main.categories c
		on c.id = pack_categories.category_id
	group by
		pc.id, v.uid
	order by
		%v
	limit %v
	offset %v;
`

var PackPageFilterOn = `
	select
	distinct
		pc.id as pack_config_id
		, pc.image_url
		, pc.item_qty
		, pc.description
		, pc.title
		, pc.vendor_id
		, pc.created_at
		, pc.token_amount
		, pc.qty
		, pc.content_main_url
		, pc.content_thumb_url
		, v.display_name as vendor_display_name
		, v.first_name as vendor_first_name
		, v.last_name as vendor_last_name
		, v.username as vendor_username
		, v.favorites_amount as vendor_favorites_amount
		, v.content_main_url as vendor_main_url
		, v.content_thumb_url as vendor_thumb_url
		, v.banner_url as vendor_banner_url
		, string_agg(c.category, ',') as raw_categories
		, count(*) over () as pack_amount
	from
		main.pack_configs pc
	join
		main.vendors v
		on v.uid = pc.vendor_id
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
	order by
		created_at desc
	limit %v
	offset %v;
`

var PackPageSortByFilterOn = `
	select
	distinct
		pc.id as pack_config_id
		, pc.image_url
		, pc.item_qty
		, pc.description
		, pc.title
		, pc.vendor_id
		, pc.created_at
		, pc.token_amount
		, pc.qty
		, pc.content_main_url
		, pc.content_thumb_url
		, v.display_name as vendor_display_name
		, v.first_name as vendor_first_name
		, v.last_name as vendor_last_name
		, v.username as vendor_username
		, v.favorites_amount as vendor_favorites_amount
		, v.content_main_url as vendor_main_url
		, v.content_thumb_url as vendor_thumb_url
		, v.banner_url as vendor_banner_url
		, string_agg(c.category, ',') as raw_categories
		, count(*) over () as pack_amount
	from
		main.pack_configs pc
	join
		main.vendors v
		on v.uid = pc.vendor_id
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
	order by
		%v 
	limit %v
	offset %v;
`

var PackItemPreview = `
	select
		pc.id as pack_config_id
		, pic.item_id
		, i.name as item_name
		, case
			when i.content_type = 'img' then 'image'
			when i.content_type = 'vid' then 'video'
		end as item_content_type
		, r.rarity as item_rarity
		, round((pic.qty::float / (pc.qty::float * pc.item_qty::float))::numeric * 100, 2) as item_chance
	from 
		main.pack_configs pc
	join
		main.pack_item_configs pic
		on pc.id = pic.pack_config_id
	join
		main.items i
		on i.id = pic.item_id
	join
		main.rarity r
		on i.rarity_id = r.id
	where 
		pc.id = %v
	order by
		r.id desc;
`
