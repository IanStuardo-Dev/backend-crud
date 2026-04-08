package transferapp

import "time"

type CreateInput struct {
	CompanyID           int64
	OriginBranchID      int64
	DestinationBranchID int64
	RequestedByUserID   int64
	Note                string
	Items               []CreateItemInput
}

type CreateItemInput struct {
	ProductID int64
	Quantity  int
}

type Output struct {
	ID                  int64
	CompanyID           int64
	OriginBranchID      int64
	DestinationBranchID int64
	Status              string
	RequestedByUserID   int64
	CompletedByUserID   int64
	Note                string
	Items               []ItemOutput
	CreatedAt           time.Time
	CompletedAt         *time.Time
}

type ItemOutput struct {
	ProductID   int64
	ProductSKU  string
	ProductName string
	Quantity    int
}
