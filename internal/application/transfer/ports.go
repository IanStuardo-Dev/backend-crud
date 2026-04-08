package transferapp

import (
	"context"

	domaintransfer "github.com/IanStuardo-Dev/backend-crud/internal/domain/transfer"
)

type Repository interface {
	Create(ctx context.Context, transfer *domaintransfer.Transfer) error
	Approve(ctx context.Context, transfer *domaintransfer.Transfer) error
	Dispatch(ctx context.Context, transfer *domaintransfer.Transfer) error
	Receive(ctx context.Context, transfer *domaintransfer.Transfer) error
	Cancel(ctx context.Context, transfer *domaintransfer.Transfer) error
	List(ctx context.Context) ([]domaintransfer.Transfer, error)
	ListByBranch(ctx context.Context, branchID int64) ([]domaintransfer.Transfer, error)
	GetByID(ctx context.Context, id int64) (*domaintransfer.Transfer, error)
}

type UseCase interface {
	Create(ctx context.Context, input CreateInput) (Output, error)
	Approve(ctx context.Context, input TransitionInput) (Output, error)
	Dispatch(ctx context.Context, input TransitionInput) (Output, error)
	Receive(ctx context.Context, input TransitionInput) (Output, error)
	Cancel(ctx context.Context, input TransitionInput) (Output, error)
	List(ctx context.Context) ([]Output, error)
	ListByBranch(ctx context.Context, branchID int64) ([]Output, error)
	GetByID(ctx context.Context, id int64) (Output, error)
}
