package transfer

import "time"

type Transfer struct {
	ID                  int64
	CompanyID           int64
	OriginBranchID      int64
	DestinationBranchID int64
	Status              string
	RequestedByUserID   int64
	CompletedByUserID   int64
	Note                string
	Items               []Item
	CreatedAt           time.Time
	CompletedAt         *time.Time
}

type Item struct {
	ProductID   int64
	ProductSKU  string
	ProductName string
	Quantity    int
}
