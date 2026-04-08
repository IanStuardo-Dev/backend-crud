package product

import "time"

const EmbeddingDimensions = 1536

// Product represents the core product entity.
type Product struct {
	ID          int64
	CompanyID   int64
	BranchID    int64
	SKU         string
	Name        string
	Description string
	Category    string
	Brand       string
	PriceCents  int64
	Currency    string
	Stock       int
	Embedding   []float32
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
