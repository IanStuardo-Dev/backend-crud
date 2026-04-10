package transferapp

import (
	transferdto "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer/dto"
	transfererrors "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer/errors"
	transferports "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer/ports"
)

type CreateInput = transferdto.CreateInput
type CreateItemInput = transferdto.CreateItemInput
type Output = transferdto.Output
type ItemOutput = transferdto.ItemOutput
type TransitionInput = transferdto.TransitionInput

type TransferWriter = transferports.TransferWriter
type TransferReader = transferports.TransferReader
type BranchReferenceReader = transferports.BranchReferenceReader
type TransferUserPolicyReader = transferports.TransferUserPolicyReader
type StockSnapshot = transferports.StockSnapshot
type InventoryTransferStockReader = transferports.InventoryTransferStockReader
type InventoryStockWriter = transferports.InventoryStockWriter
type MovementInput = transferports.MovementInput
type InventoryMovementWriter = transferports.InventoryMovementWriter
type TransactionManager = transferports.TransactionManager

var (
	ErrNotFound          = transfererrors.ErrNotFound
	ErrInvalidReference  = transfererrors.ErrInvalidReference
	ErrInsufficientStock = transfererrors.ErrInsufficientStock
	ErrForbiddenAction   = transfererrors.ErrForbiddenAction
	ErrInvalidState      = transfererrors.ErrInvalidState
)
