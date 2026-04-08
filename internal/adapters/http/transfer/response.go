package transferhttp

import "time"

type transferResponse struct {
	ID                  int64                  `json:"id"`
	CompanyID           int64                  `json:"company_id"`
	OriginBranchID      int64                  `json:"origin_branch_id"`
	DestinationBranchID int64                  `json:"destination_branch_id"`
	Status              string                 `json:"status"`
	RequestedByUserID   int64                  `json:"requested_by_user_id"`
	SupervisorUserID    int64                  `json:"supervisor_user_id"`
	ApprovedByUserID    *int64                 `json:"approved_by_user_id,omitempty"`
	DispatchedByUserID  *int64                 `json:"dispatched_by_user_id,omitempty"`
	ReceivedByUserID    *int64                 `json:"received_by_user_id,omitempty"`
	CancelledByUserID   *int64                 `json:"cancelled_by_user_id,omitempty"`
	Note                string                 `json:"note"`
	Items               []transferItemResponse `json:"items"`
	CreatedAt           time.Time              `json:"created_at"`
	ApprovedAt          *time.Time             `json:"approved_at,omitempty"`
	DispatchedAt        *time.Time             `json:"dispatched_at,omitempty"`
	ReceivedAt          *time.Time             `json:"received_at,omitempty"`
	CancelledAt         *time.Time             `json:"cancelled_at,omitempty"`
}

type transferItemResponse struct {
	ProductID   int64  `json:"product_id"`
	ProductSKU  string `json:"product_sku"`
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
}

type resourceResponse struct {
	Data any `json:"data"`
}

type collectionResponse struct {
	Data any          `json:"data"`
	Meta metaResponse `json:"meta"`
}

type metaResponse struct {
	Count int `json:"count"`
}
