package transferapp

import (
	"context"

	domaintransfer "github.com/example/crud/internal/domain/transfer"
)

type Repository interface {
	Create(ctx context.Context, transfer *domaintransfer.Transfer) error
	List(ctx context.Context) ([]domaintransfer.Transfer, error)
	GetByID(ctx context.Context, id int64) (*domaintransfer.Transfer, error)
}

type UseCase interface {
	Create(ctx context.Context, input CreateInput) (Output, error)
	List(ctx context.Context) ([]Output, error)
	GetByID(ctx context.Context, id int64) (Output, error)
}
