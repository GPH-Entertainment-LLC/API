package query

var UserFavoritePageDefault = `
	select
		f.id as favorite_id
		, f.favorited_at
		, u.uid
		, v.uid as vendor_id
		, v.created_at as vendor_created_at
		, v.updated_at as vendor_updated_at
		, v.first_name as vendor_first_name
		, v.last_name as vendor_last_name
		, v.username as vendor_username
		, v.birthday as vendor_birthday
		, v.bio as vendor_bio
		, v.active as vendor_active
		, v.display_name as vendor_display_name
		, v.favorites_amount as vendor_favorites_amount
		, v.content_thumb_url as vendor_thumb_url
		, v.content_main_url as vendor_main_url
		, v.banner_url as vendor_banner_url
		, v.pack_amount as vendor_pack_amount
		, count(*) over () as favorites_amount
	from
		main.users u
	join
		main.favorites f
		on u.uid = f.uid
		and u.uid = '%v'
	join
		main.vendors v
		on v.uid = f.vendor_id
		and v.active = true
	group by
		f.id, f.favorited_at, u.uid, v.uid
	having
		lower(v.username) like '%%%v%%'
		or lower(v.first_name) like '%%%v%%'
		or lower(v.last_name) like '%%%v%%'
		or lower(v.display_name) like '%%%v%%'
	order by
		favorited_at desc
	limit
		%v
	offset
		%v;
`
