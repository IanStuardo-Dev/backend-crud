package userhttp

type userResponse struct {
	ID              int64  `json:"id"`
	CompanyID       *int64 `json:"company_id,omitempty"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Role            string `json:"role"`
	IsActive        bool   `json:"is_active"`
	DefaultBranchID *int64 `json:"default_branch_id,omitempty"`
}

type resourceResponse struct {
	Data userResponse `json:"data"`
}

type collectionResponse struct {
	Data []userResponse `json:"data"`
	Meta metaResponse   `json:"meta"`
}

type metaResponse struct {
	Count int `json:"count"`
}
