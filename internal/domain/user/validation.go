package user

import (
	"net/mail"
	"strings"
)

// ValidationError represents a business validation failure.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}

	return e.Field + ": " + e.Message
}

func (u *User) Normalize() {
	u.Name = strings.TrimSpace(u.Name)
	u.Email = strings.TrimSpace(strings.ToLower(u.Email))
	u.Role = strings.TrimSpace(strings.ToLower(u.Role))
}

func (u User) Validate() error {
	if u.Name == "" {
		return ValidationError{Field: "name", Message: "name is required"}
	}
	if len(u.Name) < 2 {
		return ValidationError{Field: "name", Message: "name must be at least 2 characters"}
	}
	if u.Email == "" {
		return ValidationError{Field: "email", Message: "email is required"}
	}
	if _, err := mail.ParseAddress(u.Email); err != nil {
		return ValidationError{Field: "email", Message: "email format is invalid"}
	}
	if u.Role == "" {
		return ValidationError{Field: "role", Message: "role is required"}
	}
	if !isValidRole(u.Role) {
		return ValidationError{Field: "role", Message: "role is invalid"}
	}
	if u.Role != RoleSuperAdmin {
		if u.CompanyID == nil || *u.CompanyID <= 0 {
			return ValidationError{Field: "company_id", Message: "company_id is required for non-super_admin users"}
		}
	}
	if u.CompanyID != nil && *u.CompanyID <= 0 {
		return ValidationError{Field: "company_id", Message: "company_id must be greater than 0"}
	}
	if u.DefaultBranchID != nil && *u.DefaultBranchID <= 0 {
		return ValidationError{Field: "default_branch_id", Message: "default_branch_id must be greater than 0"}
	}

	return nil
}

func isValidRole(role string) bool {
	switch role {
	case RoleSuperAdmin, RoleCompanyAdmin, RoleInventoryManager, RoleSalesUser:
		return true
	default:
		return false
	}
}

func ValidatePassword(password string) error {
	if strings.TrimSpace(password) == "" {
		return ValidationError{Field: "password", Message: "password is required"}
	}
	if len(password) < 8 {
		return ValidationError{Field: "password", Message: "password must be at least 8 characters"}
	}

	var hasLetter bool
	var hasDigit bool
	for _, char := range password {
		if ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z') {
			hasLetter = true
		}
		if '0' <= char && char <= '9' {
			hasDigit = true
		}
	}

	if !hasLetter || !hasDigit {
		return ValidationError{Field: "password", Message: "password must contain at least one letter and one number"}
	}

	return nil
}
