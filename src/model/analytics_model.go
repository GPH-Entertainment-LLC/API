package model

type CustomerAnalytic struct {
	Username       *string  `db:"username" json:"username"`
	PacksPurchased *uint64  `db:"packs_purchased" json:"packsPurchased"`
	AmountSpent    *float64 `db:"amount_spent" json:"amountSpent"`
}

type PackSalesAnalytic struct {
	Granularity *string  `db:"granularity" json:"granularity"`
	QtySold     *uint64  `db:"qty_sold" json:"qtySold"`
	TotalSales  *float64 `db:"total_sales" json:"totalSales"`
}

type PackQtySoldAnalytic struct {
	PackTitle *string `db:"pack_title" json:"packTitle"`
	QtySold   *uint64 `db:"qty_sold" json:"qtySold"`
}

type PackAnalyticsResp struct {
	TimeAxis []string             `json:"timeAxis"`
	DataSet  []*PackSalesAnalytic `json:"dataSet"`
}
