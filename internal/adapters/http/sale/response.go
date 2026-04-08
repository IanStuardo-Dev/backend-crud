package salehttp

import "time"

type saleResponse struct {
	ID               int64              `json:"id"`
	CompanyID        int64              `json:"company_id"`
	BranchID         int64              `json:"branch_id"`
	CreatedByUserID  int64              `json:"created_by_user_id"`
	TotalAmountCents int64              `json:"total_amount_cents"`
	Items            []saleItemResponse `json:"items"`
	CreatedAt        time.Time          `json:"created_at"`
}

type saleItemResponse struct {
	ProductID      int64  `json:"product_id"`
	ProductSKU     string `json:"product_sku"`
	ProductName    string `json:"product_name"`
	Quantity       int    `json:"quantity"`
	UnitPriceCents int64  `json:"unit_price_cents"`
	SubtotalCents  int64  `json:"subtotal_cents"`
}

type resourceResponse struct {
	Data saleResponse `json:"data"`
}

type collectionResponse struct {
	Data []saleResponse `json:"data"`
	Meta metaResponse   `json:"meta"`
}

type metaResponse struct {
	Count int `json:"count"`
}
