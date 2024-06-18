package query

// TODO -> update to default vendor favorites amount desc
var VendorPageDefault = `
	select
		v.uid
		, v.created_at
		, v.updated_at
		, v.image_url
		, v.first_name
		, v.last_name
		, v.display_name
		, v.username
		, v.birthday
		, v.bio
		, v.is_vendor
		, v.active
		, v.favorites_amount
		, v.content_main_url
		, v.content_thumb_url
		, v.banner_url
		, v.pack_amount
		, count(*) over () as vendor_amount
		, string_agg(c.category, ',') as categories
	from
		main.vendors v
	left join
		main.vendor_categories vc on vc.vendor_id = v.uid
	left join
		main.categories c on vc.category_id = c.id
	where
		v.active = true
	group by
		v.uid
	having
		lower(v.username) like '%%%v%%'
		or lower(v.display_name) like '%%%v%%'
	Order By v.created_at desc
	Limit %v
	Offset %v;
`

var VendorPageSortBy = `
	select
		v.uid
		, v.created_at
		, v.updated_at
		, v.image_url
		, v.first_name
		, v.last_name
		, v.display_name
		, v.username
		, v.birthday
		, v.bio
		, v.is_vendor
		, v.active
		, v.favorites_amount
		, v.content_main_url
		, v.content_thumb_url
		, v.banner_url
		, v.pack_amount
		, count(*) over () as vendor_amount
		, string_agg(c.category, ',') as categories
	from
		main.vendors v
	left join
		main.vendor_categories vc on vc.vendor_id = v.uid
	left join
		main.categories c on vc.category_id = c.id
	where
		v.active = true
	Group By
		v.uid
	having
		lower(v.username) like '%%%v%%'
		or lower(v.display_name) like '%%%v%%'
	Order By %v
	Limit %v
	Offset %v;
`

var VendorPageFilterOn = `
	select
		v.uid
		, v.created_at
		, v.updated_at
		, v.image_url
		, v.first_name
		, v.last_name
		, v.display_name
		, v.username
		, v.birthday
		, v.bio
		, v.is_vendor
		, v.active
		, v.favorites_amount
		, v.content_main_url
		, v.content_thumb_url
		, v.banner_url
		, v.pack_amount
		, count(*) over () as vendor_amount
		, string_agg(c.category, ',') categories
	from
		main.vendors v
	left join
		main.vendor_categories vc on vc.vendor_id = v.uid
	left join
		main.categories c on vc.category_id = c.id
	where
		v.active = true
	Group By
		v.uid
	having
		string_agg(c.category, ',') like '%%%v%%'
		and (
			lower(v.username) like '%%%v%%'
			or lower(v.display_name) like '%%%v%%'
		)
	Order By v.created_at desc
	Limit %v
	Offset %v;
`

var VendorPageSortByFilterOn = `
	select
		v.uid
		, v.created_at
		, v.updated_at
		, v.image_url
		, v.first_name
		, v.last_name
		, v.display_name
		, v.username
		, v.birthday
		, v.bio
		, v.is_vendor
		, v.active
		, v.favorites_amount
		, v.content_main_url
		, v.content_thumb_url
		, v.banner_url
		, v.pack_amount
		, count(*) over () as vendor_amount
		, string_agg(c.category, ',') categories
	from
		main.vendors v
	left join
		main.vendor_categories vc on vc.vendor_id = v.uid
	left join
		main.categories c on vc.category_id = c.id
	where
		v.active = true
	Group By
		v.uid
	having
		string_agg(c.category, ',') like '%%%v%%'
		and (
			lower(v.username) like '%%%v%%'
			or lower(v.display_name) like '%%%v%%'
		)
	Order By %v
	Limit %v
	Offset %v;
`
