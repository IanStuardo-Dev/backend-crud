package inventoryapp

type ListByBranchInput struct {
	CompanyID int64
	BranchID  int64
}

type SuggestSourcesInput struct {
	CompanyID           int64
	DestinationBranchID int64
	ProductID           int64
	Quantity            int
}

type ItemOutput struct {
	CompanyID      int64
	BranchID       int64
	ProductID      int64
	ProductSKU     string
	ProductName    string
	Category       string
	Brand          string
	StockOnHand    int
	ReservedStock  int
	AvailableStock int
}

type SourceCandidateOutput struct {
	CompanyID          int64
	BranchID           int64
	BranchCode         string
	BranchName         string
	City               string
	Region             string
	Latitude           *float64
	Longitude          *float64
	ProductID          int64
	AvailableStock     int
	DistanceKilometers *float64
}
