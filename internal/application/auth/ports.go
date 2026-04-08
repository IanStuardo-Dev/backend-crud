package authapp

import (
	"context"

	domainuser "github.com/example/crud/internal/domain/user"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*domainuser.User, error)
}

type PasswordVerifier interface {
	Compare(hash, password string) error
}

type TokenManager interface {
	Generate(user AuthenticatedUser) (IssuedToken, error)
	Verify(token string) (AuthenticatedUser, error)
}

type UseCase interface {
	Login(ctx context.Context, input LoginInput) (LoginOutput, error)
}
