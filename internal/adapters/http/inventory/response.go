package inventoryhttp

type itemResponse struct {
	CompanyID      int64  `json:"company_id"`
	BranchID       int64  `json:"branch_id"`
	ProductID      int64  `json:"product_id"`
	ProductSKU     string `json:"product_sku"`
	ProductName    string `json:"product_name"`
	Category       string `json:"category"`
	Brand          string `json:"brand"`
	StockOnHand    int    `json:"stock_on_hand"`
	ReservedStock  int    `json:"reserved_stock"`
	AvailableStock int    `json:"available_stock"`
}

type sourceCandidateResponse struct {
	CompanyID          int64    `json:"company_id"`
	BranchID           int64    `json:"branch_id"`
	BranchCode         string   `json:"branch_code"`
	BranchName         string   `json:"branch_name"`
	City               string   `json:"city"`
	Region             string   `json:"region"`
	Latitude           *float64 `json:"latitude,omitempty"`
	Longitude          *float64 `json:"longitude,omitempty"`
	ProductID          int64    `json:"product_id"`
	AvailableStock     int      `json:"available_stock"`
	DistanceKilometers *float64 `json:"distance_kilometers,omitempty"`
}

type collectionResponse struct {
	Data any          `json:"data"`
	Meta metaResponse `json:"meta"`
}

type metaResponse struct {
	Count int `json:"count"`
}
