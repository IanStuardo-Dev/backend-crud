package salehttp

type createSaleRequest struct {
	CompanyID int64                   `json:"company_id"`
	BranchID  int64                   `json:"branch_id"`
	Items     []createSaleItemRequest `json:"items"`
}

type createSaleItemRequest struct {
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}
