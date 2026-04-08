package transferhttp

type createTransferRequest struct {
	CompanyID           int64                       `json:"company_id"`
	OriginBranchID      int64                       `json:"origin_branch_id"`
	DestinationBranchID int64                       `json:"destination_branch_id"`
	SupervisorUserID    int64                       `json:"supervisor_user_id"`
	Note                string                      `json:"note"`
	Items               []createTransferItemRequest `json:"items"`
}

type createTransferItemRequest struct {
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}
