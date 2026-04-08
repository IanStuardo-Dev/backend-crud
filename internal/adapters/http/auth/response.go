package authhttp

import "time"

type loginResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresAt   time.Time    `json:"expires_at"`
	User        userResponse `json:"user"`
}

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
	Data any `json:"data"`
}
