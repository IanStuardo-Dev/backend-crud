package saleapp

import (
	"context"

	salecommand "github.com/IanStuardo-Dev/backend-crud/internal/application/sale/command"
	salequery "github.com/IanStuardo-Dev/backend-crud/internal/application/sale/query"
)

type UseCase interface {
	Create(ctx context.Context, input CreateInput) (Output, error)
	List(ctx context.Context) ([]Output, error)
	GetByID(ctx context.Context, id int64) (Output, error)
}

type useCase struct {
	create  salecommand.CreateHandler
	list    salequery.ListHandler
	getByID salequery.GetByIDHandler
}

type inlineTransactionManager struct{}

func (inlineTransactionManager) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func newUseCase(
	txManager TransactionManager,
	saleReader SaleReader,
	saleWriter SaleWriter,
	stockReader InventoryStockReader,
	stockWriter InventoryStockWriter,
	movementWriter InventoryMovementWriter,
	branchReader BranchReferenceReader,
	userReader UserReferenceReader,
) UseCase {
	return &useCase{
		create:  salecommand.NewCreateHandler(txManager, saleWriter, stockReader, stockWriter, movementWriter, branchReader, userReader),
		list:    salequery.NewListHandler(saleReader),
		getByID: salequery.NewGetByIDHandler(saleReader),
	}
}

func NewUseCase(args ...any) UseCase {
	switch len(args) {
	case 1:
		repo := args[0]
		return newUseCase(
			inlineTransactionManager{},
			repo.(SaleReader),
			repo.(SaleWriter),
			repo.(InventoryStockReader),
			repo.(InventoryStockWriter),
			repo.(InventoryMovementWriter),
			repo.(BranchReferenceReader),
			repo.(UserReferenceReader),
		)
	case 8:
		return newUseCase(
			args[0].(TransactionManager),
			args[1].(SaleReader),
			args[2].(SaleWriter),
			args[3].(InventoryStockReader),
			args[4].(InventoryStockWriter),
			args[5].(InventoryMovementWriter),
			args[6].(BranchReferenceReader),
			args[7].(UserReferenceReader),
		)
	default:
		panic("invalid sale use case dependencies")
	}
}

func (uc *useCase) Create(ctx context.Context, input CreateInput) (Output, error) {
	return uc.create.Handle(ctx, input)
}

func (uc *useCase) List(ctx context.Context) ([]Output, error) {
	return uc.list.Handle(ctx)
}

func (uc *useCase) GetByID(ctx context.Context, id int64) (Output, error) {
	return uc.getByID.Handle(ctx, id)
}
