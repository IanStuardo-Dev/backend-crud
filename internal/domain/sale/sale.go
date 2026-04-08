package sale

import "time"

// Sale represents a stock-discounting commercial transaction.
type Sale struct {
	ID               int64
	CompanyID        int64
	BranchID         int64
	CreatedByUserID  int64
	TotalAmountCents int64
	Items            []Item
	CreatedAt        time.Time
}

// Item represents a sold product and its inventory impact.
type Item struct {
	ProductID      int64
	ProductSKU     string
	ProductName    string
	Quantity       int
	UnitPriceCents int64
	SubtotalCents  int64
}
