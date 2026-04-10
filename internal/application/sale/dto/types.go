package saledto

import "time"

type CreateInput struct {
	CompanyID       int64
	BranchID        int64
	CreatedByUserID int64
	Items           []CreateItemInput
}

type CreateItemInput struct {
	ProductID int64
	Quantity  int
}

type Output struct {
	ID               int64
	CompanyID        int64
	BranchID         int64
	CreatedByUserID  int64
	TotalAmountCents int64
	Items            []ItemOutput
	CreatedAt        time.Time
}

type ItemOutput struct {
	ProductID      int64
	ProductSKU     string
	ProductName    string
	Quantity       int
	UnitPriceCents int64
	SubtotalCents  int64
}
