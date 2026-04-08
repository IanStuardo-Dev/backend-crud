package userhttp

type createUserRequest struct {
	CompanyID       *int64 `json:"company_id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Role            string `json:"role"`
	IsActive        *bool  `json:"is_active"`
	DefaultBranchID *int64 `json:"default_branch_id"`
	Password        string `json:"password"`
}

type updateUserRequest struct {
	CompanyID       *int64 `json:"company_id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Role            string `json:"role"`
	IsActive        bool   `json:"is_active"`
	DefaultBranchID *int64 `json:"default_branch_id"`
}
