package inventoryhttp

type listByBranchRequest struct {
	CompanyID int64
	BranchID  int64
}

type suggestSourcesRequest struct {
	CompanyID           int64
	DestinationBranchID int64
	ProductID           int64
	Quantity            int
}
