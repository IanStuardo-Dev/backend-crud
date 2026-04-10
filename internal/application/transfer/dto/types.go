package transferdto

import "time"

type CreateInput struct {
	CompanyID           int64
	OriginBranchID      int64
	DestinationBranchID int64
	RequestedByUserID   int64
	SupervisorUserID    int64
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
	SupervisorUserID    int64
	ApprovedByUserID    *int64
	DispatchedByUserID  *int64
	ReceivedByUserID    *int64
	CancelledByUserID   *int64
	Note                string
	Items               []ItemOutput
	CreatedAt           time.Time
	ApprovedAt          *time.Time
	DispatchedAt        *time.Time
	ReceivedAt          *time.Time
	CancelledAt         *time.Time
}

type ItemOutput struct {
	ProductID   int64
	ProductSKU  string
	ProductName string
	Quantity    int
}

type TransitionInput struct {
	ID          int64
	ActorUserID int64
}
