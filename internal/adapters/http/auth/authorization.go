package authhttp

import (
	"context"

	authapp "github.com/IanStuardo-Dev/backend-crud/internal/application/auth"
)

func EnsureCompanyAccess(ctx context.Context, companyID int64) error {
	user, ok := AuthenticatedUserFromContext(ctx)
	if !ok {
		return nil
	}
	if user.CanAccessCompany(companyID) {
		return nil
	}
	return authapp.ErrForbidden
}
