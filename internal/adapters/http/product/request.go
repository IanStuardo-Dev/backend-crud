package producthttp

type createProductRequest struct {
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
}

type updateProductRequest struct {
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
}
