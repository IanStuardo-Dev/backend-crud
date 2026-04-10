package productdto

import "time"

type CreateInput struct {
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
}

type UpdateInput struct {
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
}

type Output struct {
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

type FindNeighborsInput struct {
	ProductID     int64
	Limit         int
	MinSimilarity float64
}

type NeighborOutput struct {
	ProductID            int64
	SKU                  string
	Name                 string
	Description          string
	Category             string
	Brand                string
	PriceCents           int64
	Currency             string
	SimilarityPercentage float64
	Distance             float64
}

type FindNeighborsOutput struct {
	SourceProductID   int64
	SourceProductName string
	SourceCompanyID   int64
	Neighbors         []NeighborOutput
	Limit             int
}

type RecordNeighborFeedbackInput struct {
	SourceProductID    int64
	SuggestedProductID int64
	CompanyID          int64
	BranchID           int64
	UserID             int64
	Action             string
	Note               string
}

type NeighborFeedbackOutput struct {
	SourceProductID    int64
	SuggestedProductID int64
	CompanyID          int64
	BranchID           int64
	UserID             int64
	Action             string
	Note               string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
