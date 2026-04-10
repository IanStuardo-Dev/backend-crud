package transferapp

import (
	"context"

	transfercommand "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer/command"
	transferquery "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer/query"
)

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

type useCase struct {
	create       transfercommand.CreateHandler
	transition   transfercommand.TransitionHandler
	list         transferquery.ListHandler
	listByBranch transferquery.ListByBranchHandler
	getByID      transferquery.GetByIDHandler
}

type inlineTransactionManager struct{}

func (inlineTransactionManager) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func newUseCase(
	txManager TransactionManager,
	reader TransferReader,
	writer TransferWriter,
	stockReader InventoryTransferStockReader,
	stockWriter InventoryStockWriter,
	movementWriter InventoryMovementWriter,
	branchReader BranchReferenceReader,
	userPolicies TransferUserPolicyReader,
) UseCase {
	return &useCase{
		create:       transfercommand.NewCreateHandler(txManager, writer, stockReader, branchReader, userPolicies),
		transition:   transfercommand.NewTransitionHandler(txManager, reader, writer, stockReader, stockWriter, movementWriter),
		list:         transferquery.NewListHandler(reader),
		listByBranch: transferquery.NewListByBranchHandler(reader),
		getByID:      transferquery.NewGetByIDHandler(reader),
	}
}

func NewUseCase(args ...any) UseCase {
	switch len(args) {
	case 1:
		repo := args[0]
		return newUseCase(
			inlineTransactionManager{},
			repo.(TransferReader),
			repo.(TransferWriter),
			repo.(InventoryTransferStockReader),
			repo.(InventoryStockWriter),
			repo.(InventoryMovementWriter),
			repo.(BranchReferenceReader),
			repo.(TransferUserPolicyReader),
		)
	case 8:
		return newUseCase(
			args[0].(TransactionManager),
			args[1].(TransferReader),
			args[2].(TransferWriter),
			args[3].(InventoryTransferStockReader),
			args[4].(InventoryStockWriter),
			args[5].(InventoryMovementWriter),
			args[6].(BranchReferenceReader),
			args[7].(TransferUserPolicyReader),
		)
	default:
		panic("invalid transfer use case dependencies")
	}
}

func (uc *useCase) Create(ctx context.Context, input CreateInput) (Output, error) {
	return uc.create.Handle(ctx, input)
}

func (uc *useCase) Approve(ctx context.Context, input TransitionInput) (Output, error) {
	return uc.transition.Approve(ctx, input)
}

func (uc *useCase) Dispatch(ctx context.Context, input TransitionInput) (Output, error) {
	return uc.transition.Dispatch(ctx, input)
}

func (uc *useCase) Receive(ctx context.Context, input TransitionInput) (Output, error) {
	return uc.transition.Receive(ctx, input)
}

func (uc *useCase) Cancel(ctx context.Context, input TransitionInput) (Output, error) {
	return uc.transition.Cancel(ctx, input)
}

func (uc *useCase) List(ctx context.Context) ([]Output, error) {
	return uc.list.Handle(ctx)
}

func (uc *useCase) ListByBranch(ctx context.Context, branchID int64) ([]Output, error) {
	return uc.listByBranch.Handle(ctx, branchID)
}

func (uc *useCase) GetByID(ctx context.Context, id int64) (Output, error) {
	return uc.getByID.Handle(ctx, id)
}
