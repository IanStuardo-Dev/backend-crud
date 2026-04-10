package transferports

import (
	"context"

	apptx "github.com/IanStuardo-Dev/backend-crud/internal/application/shared/tx"
	domaintransfer "github.com/IanStuardo-Dev/backend-crud/internal/domain/transfer"
)

type TransferWriter interface {
	Create(ctx context.Context, transfer *domaintransfer.Transfer) error
	Approve(ctx context.Context, transfer *domaintransfer.Transfer) error
	Dispatch(ctx context.Context, transfer *domaintransfer.Transfer) error
	Receive(ctx context.Context, transfer *domaintransfer.Transfer) error
	Cancel(ctx context.Context, transfer *domaintransfer.Transfer) error
}

type TransferReader interface {
	List(ctx context.Context) ([]domaintransfer.Transfer, error)
	ListByBranch(ctx context.Context, branchID int64) ([]domaintransfer.Transfer, error)
	GetByID(ctx context.Context, id int64) (*domaintransfer.Transfer, error)
}

type BranchReferenceReader interface {
	BranchExists(ctx context.Context, companyID, branchID int64) (bool, error)
}

type TransferUserPolicyReader interface {
	RequesterCanAct(ctx context.Context, companyID, userID int64) (bool, error)
	SupervisorEligible(ctx context.Context, companyID, userID int64) (bool, error)
}

type StockSnapshot struct {
	ProductID      int64
	ProductSKU     string
	ProductName    string
	StockOnHand    int
	ReservedStock  int
	AvailableStock int
}

type InventoryTransferStockReader interface {
	LoadForTransfer(ctx context.Context, companyID, branchID int64, items []domaintransfer.Item, lock bool) (map[int64]StockSnapshot, error)
}

type InventoryStockWriter interface {
	PutStockOnHand(ctx context.Context, companyID, branchID, productID int64, stockOnHand int) error
}

type MovementInput struct {
	CompanyID       int64
	BranchID        int64
	ProductID       int64
	TransferID      int64
	MovementType    string
	QuantityDelta   int
	StockAfter      int
	CreatedByUserID int64
}

type InventoryMovementWriter interface {
	CreateTransferMovement(ctx context.Context, input MovementInput) error
}

type TransactionManager = apptx.Manager
