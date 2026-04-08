package producthttp

import "time"

type productResponse struct {
	ID          int64     `json:"id"`
	CompanyID   int64     `json:"company_id"`
	BranchID    int64     `json:"branch_id"`
	SKU         string    `json:"sku"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Brand       string    `json:"brand"`
	PriceCents  int64     `json:"price_cents"`
	Currency    string    `json:"currency"`
	Stock       int       `json:"stock"`
	Embedding   []float32 `json:"embedding,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type resourceResponse struct {
	Data productResponse `json:"data"`
}

type collectionResponse struct {
	Data []productResponse `json:"data"`
	Meta metaResponse      `json:"meta"`
}

type neighborsCollectionResponse struct {
	Data []neighborResponse    `json:"data"`
	Meta neighborsMetaResponse `json:"meta"`
}

type neighborResponse struct {
	ProductID            int64   `json:"product_id"`
	SKU                  string  `json:"sku"`
	Name                 string  `json:"name"`
	Description          string  `json:"description"`
	Category             string  `json:"category"`
	Brand                string  `json:"brand"`
	PriceCents           int64   `json:"price_cents"`
	Currency             string  `json:"currency"`
	SimilarityPercentage float64 `json:"similarity_percentage"`
	Distance             float64 `json:"distance"`
}

type neighborsMetaResponse struct {
	SourceProductID   int64  `json:"source_product_id"`
	SourceProductName string `json:"source_product_name"`
	Count             int    `json:"count"`
	Limit             int    `json:"limit"`
}

type metaResponse struct {
	Count int `json:"count"`
}
