package authapp

import "time"

const roleSuperAdmin = "super_admin"

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	AccessToken string
	TokenType   string
	ExpiresAt   time.Time
	User        UserOutput
}

type UserOutput struct {
	ID              int64
	CompanyID       *int64
	Name            string
	Email           string
	Role            string
	IsActive        bool
	DefaultBranchID *int64
}

type AuthenticatedUser struct {
	ID              int64
	CompanyID       *int64
	Name            string
	Email           string
	Role            string
	IsActive        bool
	DefaultBranchID *int64
}

type IssuedToken struct {
	AccessToken string
	ExpiresAt   time.Time
}

func (u AuthenticatedUser) IsSuperAdmin() bool {
	return u.Role == roleSuperAdmin
}

func (u AuthenticatedUser) CanAccessCompany(companyID int64) bool {
	if u.IsSuperAdmin() {
		return true
	}
	return u.CompanyID != nil && *u.CompanyID == companyID
}

func (u AuthenticatedUser) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if u.Role == role {
			return true
		}
	}
	return false
}
