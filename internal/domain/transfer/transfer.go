package transfer

import "time"

const (
	StatusPendingApproval = "pending_approval"
	StatusApproved        = "approved"
	StatusInTransit       = "in_transit"
	StatusReceived        = "received"
	StatusCancelled       = "cancelled"
)

type Transfer struct {
	ID                  int64
	CompanyID           int64
	OriginBranchID      int64
	DestinationBranchID int64
	Status              string
	RequestedByUserID   int64
	SupervisorUserID    int64
	ApprovedByUserID    *int64
	DispatchedByUserID  *int64
	ReceivedByUserID    *int64
	CancelledByUserID   *int64
	Note                string
	Items               []Item
	CreatedAt           time.Time
	ApprovedAt          *time.Time
	DispatchedAt        *time.Time
	ReceivedAt          *time.Time
	CancelledAt         *time.Time
}

type Item struct {
	ProductID   int64
	ProductSKU  string
	ProductName string
	Quantity    int
}
