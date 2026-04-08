package userapp

import (
	"context"

	domainuser "github.com/example/crud/internal/domain/user"
)

// Repository defines the persistence contract required by the use case.
type Repository interface {
	Create(ctx context.Context, user *domainuser.User) error
	List(ctx context.Context) ([]domainuser.User, error)
	GetByID(ctx context.Context, id int64) (*domainuser.User, error)
	GetByEmail(ctx context.Context, email string) (*domainuser.User, error)
	Update(ctx context.Context, user *domainuser.User) error
	Delete(ctx context.Context, id int64) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
}

// UseCase defines the application operations for users.
type UseCase interface {
	Create(ctx context.Context, input CreateInput) (Output, error)
	List(ctx context.Context) ([]Output, error)
	GetByID(ctx context.Context, id int64) (Output, error)
	Update(ctx context.Context, input UpdateInput) (Output, error)
	Delete(ctx context.Context, id int64) error
}
