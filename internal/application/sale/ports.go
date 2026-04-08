package saleapp

import (
	"context"

	domainsale "github.com/IanStuardo-Dev/backend-crud/internal/domain/sale"
)

// Repository defines the persistence contract required by the sales use case.
type Repository interface {
	Create(ctx context.Context, sale *domainsale.Sale) error
	List(ctx context.Context) ([]domainsale.Sale, error)
	GetByID(ctx context.Context, id int64) (*domainsale.Sale, error)
}

// UseCase defines the application operations for sales.
type UseCase interface {
	Create(ctx context.Context, input CreateInput) (Output, error)
	List(ctx context.Context) ([]Output, error)
	GetByID(ctx context.Context, id int64) (Output, error)
}
