package transferhttp

import "time"

type transferResponse struct {
	ID                  int64                  `json:"id"`
	CompanyID           int64                  `json:"company_id"`
	OriginBranchID      int64                  `json:"origin_branch_id"`
	DestinationBranchID int64                  `json:"destination_branch_id"`
	Status              string                 `json:"status"`
	RequestedByUserID   int64                  `json:"requested_by_user_id"`
	CompletedByUserID   int64                  `json:"completed_by_user_id"`
	Note                string                 `json:"note"`
	Items               []transferItemResponse `json:"items"`
	CreatedAt           time.Time              `json:"created_at"`
	CompletedAt         *time.Time             `json:"completed_at,omitempty"`
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
