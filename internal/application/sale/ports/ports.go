package saleports

import (
	"context"

	saledto "github.com/IanStuardo-Dev/backend-crud/internal/application/sale/dto"
	apptx "github.com/IanStuardo-Dev/backend-crud/internal/application/shared/tx"
	domainsale "github.com/IanStuardo-Dev/backend-crud/internal/domain/sale"
)

type SaleWriter interface {
	Create(ctx context.Context, sale *domainsale.Sale) error
}

type SaleReader interface {
	List(ctx context.Context) ([]domainsale.Sale, error)
	GetByID(ctx context.Context, id int64) (*domainsale.Sale, error)
}

type BranchReferenceReader interface {
	BranchExists(ctx context.Context, companyID, branchID int64) (bool, error)
}

type UserReferenceReader interface {
	UserExists(ctx context.Context, userID int64) (bool, error)
}

type StockSnapshot struct {
	ProductID      int64
	SKU            string
	Name           string
	PriceCents     int64
	StockOnHand    int
	ReservedStock  int
	AvailableStock int
}

type InventoryStockReader interface {
	LoadForSale(ctx context.Context, companyID, branchID int64, items []domainsale.Item, lock bool) (map[int64]StockSnapshot, error)
}

type InventoryStockWriter interface {
	SetStockOnHand(ctx context.Context, companyID, branchID, productID int64, stockOnHand int) error
}

type MovementInput struct {
	CompanyID       int64
	BranchID        int64
	ProductID       int64
	SaleID          int64
	QuantityDelta   int
	StockAfter      int
	CreatedByUserID int64
}

type InventoryMovementWriter interface {
	CreateSaleMovement(ctx context.Context, input MovementInput) error
}

type TransactionManager = apptx.Manager

type OutputMapper interface {
	ToOutput(sale domainsale.Sale) saledto.Output
}
