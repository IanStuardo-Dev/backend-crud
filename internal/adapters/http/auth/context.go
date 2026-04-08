package authhttp

import (
	"context"

	authapp "github.com/IanStuardo-Dev/backend-crud/internal/application/auth"
)

type contextKey string

const authenticatedUserKey contextKey = "authenticated-user"

func withAuthenticatedUser(ctx context.Context, user authapp.AuthenticatedUser) context.Context {
	return context.WithValue(ctx, authenticatedUserKey, user)
}

func AuthenticatedUserFromContext(ctx context.Context) (authapp.AuthenticatedUser, bool) {
	user, ok := ctx.Value(authenticatedUserKey).(authapp.AuthenticatedUser)
	return user, ok
}
