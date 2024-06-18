package query

var UserPackPageDefault = `
select
	distinct
		pc.id as pack_config_id
		, pc.image_url
		, pc.item_qty
		, pc.description
		, pc.title
		, pc.created_at
		, pc.token_amount
		, pc.qty
		, pc.content_main_url
		, pc.content_thumb_url
		, string_agg(c.category, ',') as raw_categories
		, p.raw_pack_ids
		, p.pack_amount
		, p.latest_purchased_at
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
	join (
		select
			pack_config_id
			, string_agg(id::text, ',') as raw_pack_ids
			, count(pack_config_id) as pack_amount
			, max(purchased_at) as latest_purchased_at
		from main.pack_facts
		where owner_id = '%v'
		and opened_at is null
		group by pack_config_id
	) p
		on pc.id = p.pack_config_id
	join
		main.vendors v
		on pc.vendor_id = v.uid
	left join
		main.pack_categories
		on pack_categories.pack_config_id = pc.id
	left join
		main.categories c
		on c.id = pack_categories.category_id
	group by
		pc.id, v.uid, p.raw_pack_ids, p.pack_amount, p.latest_purchased_at
	having
		lower(v.username) like '%%%v%%'
		or lower(v.first_name) like '%%%v%%'
		or lower(v.last_name) like '%%%v%%'
		or lower(v.display_name) like '%%%v%%'
		or lower(pc.title) like '%%%v%%'
	order by
		latest_purchased_at desc
	limit %v
	offset %v	
`

var UserPackPageSortBy = `
select
	distinct
		pc.id as pack_config_id
		, pc.image_url
		, pc.item_qty
		, pc.description
		, pc.title
		, pc.created_at
		, pc.token_amount
		, pc.qty
		, pc.content_main_url
		, pc.content_thumb_url
		, string_agg(c.category, ',') as raw_categories
		, p.raw_pack_ids
		, p.pack_amount
		, p.latest_purchased_at
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
	join (
		select
			pack_config_id
			, string_agg(id::text, ',') as raw_pack_ids
			, count(pack_config_id) as pack_amount
			, max(purchased_at) as latest_purchased_at
		from main.pack_facts
		where owner_id = '%v'
		and opened_at is null
		group by pack_config_id
	) p
		on pc.id = p.pack_config_id
	join
		main.vendors v
		on pc.vendor_id = v.uid
	left join
		main.pack_categories
		on pack_categories.pack_config_id = pc.id
	left join
		main.categories c
		on c.id = pack_categories.category_id
	group by
		pc.id, v.uid, p.raw_pack_ids, p.pack_amount, p.latest_purchased_at
	having
		lower(v.username) like '%%%v%%'
		or lower(v.first_name) like '%%%v%%'
		or lower(v.last_name) like '%%%v%%'
		or lower(v.display_name) like '%%%v%%'
		or lower(pc.title) like '%%%v%%'
	order by
		%v
	limit %v
	offset %v	
`

var UserPackPageFilterOn = `
select
	distinct
		pc.id as pack_config_id
		, pc.image_url
		, pc.item_qty
		, pc.description
		, pc.title
		, pc.created_at
		, pc.token_amount
		, pc.qty
		, pc.content_main_url
		, pc.content_thumb_url
		, string_agg(c.category, ',') as raw_categories
		, p.raw_pack_ids
		, p.pack_amount
		, p.latest_purchased_at
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
	join (
		select
			pack_config_id
			, string_agg(id::text, ',') as raw_pack_ids
			, count(pack_config_id) as pack_amount
			, max(purchased_at) as latest_purchased_at
		from main.pack_facts
		where owner_id = '%v'
		and opened_at is null
		group by pack_config_id
	) p
		on pc.id = p.pack_config_id
	join
		main.vendors v
		on pc.vendor_id = v.uid
	left join
		main.pack_categories
		on pack_categories.pack_config_id = pc.id
	left join
		main.categories c
		on c.id = pack_categories.category_id
	group by
		pc.id, v.uid, p.raw_pack_ids, p.pack_amount, p.latest_purchased_at
	having
		string_agg(c.category, ',') like '%%%v%%'
		and (lower(v.username) like '%%%v%%'
			or lower(v.first_name) like '%%%v%%'
			or lower(v.last_name) like '%%%v%%'
			or lower(v.display_name) like '%%%v%%'
			or lower(pc.title) like '%%%v%%'
		)
	order by
		latest_purchased_at desc
	limit %v
	offset %v
`

var UserPackPageSortByFilterOn = `
	select
		distinct
			pc.id as pack_config_id
			, pc.image_url
			, pc.item_qty
			, pc.description
			, pc.title
			, pc.created_at
			, pc.token_amount
			, pc.qty
			, pc.content_main_url
			, pc.content_thumb_url
			, string_agg(c.category, ',') as raw_categories
			, p.raw_pack_ids
			, p.pack_amount
			, p.latest_purchased_at
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
		join (
			select
				pack_config_id
				, string_agg(id::text, ',') as raw_pack_ids
				, count(pack_config_id) as pack_amount
				, max(purchased_at) as latest_purchased_at
			from main.pack_facts
			where owner_id = '%v'
			and opened_at is null
			group by pack_config_id
		) p
			on pc.id = p.pack_config_id
		join
			main.vendors v
			on pc.vendor_id = v.uid
		left join
			main.pack_categories
			on pack_categories.pack_config_id = pc.id
		left join
			main.categories c
			on c.id = pack_categories.category_id
		group by
			pc.id, v.uid, p.raw_pack_ids, p.pack_amount, p.latest_purchased_at
		having
			string_agg(c.category, ',') like '%%%v%%'
			and (
				lower(v.username) like '%%%v%%'
				or lower(v.first_name) like '%%%v%%'
				or lower(v.last_name) like '%%%v%%'
				or lower(v.display_name) like '%%%v%%'
				or lower(pc.title) like '%%%v%%'
			)
		order by
			%v
		limit %v
		offset %v	
`
