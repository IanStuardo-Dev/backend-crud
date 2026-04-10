package transfercommand

import (
	"context"

	transfererrors "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer/errors"
	transferports "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer/ports"
	domaintransfer "github.com/IanStuardo-Dev/backend-crud/internal/domain/transfer"
)

func (h TransitionHandler) persistDispatch(ctx context.Context, transfer *domaintransfer.Transfer) error {
	productRows, err := h.stockReader.LoadForTransfer(ctx, transfer.CompanyID, transfer.OriginBranchID, transfer.Items, true)
	if err != nil {
		return err
	}

	for _, item := range transfer.Items {
		originRow, ok := productRows[item.ProductID]
		if !ok || originRow.ProductSKU == "" {
			return transfererrors.ErrInvalidReference
		}
		if originRow.AvailableStock < item.Quantity {
			return transfererrors.ErrInsufficientStock
		}

		stockAfter := originRow.StockOnHand - item.Quantity
		if err := h.stockWriter.PutStockOnHand(ctx, transfer.CompanyID, transfer.OriginBranchID, item.ProductID, stockAfter); err != nil {
			return err
		}
		if err := h.movementWriter.CreateTransferMovement(ctx, transferports.MovementInput{
			CompanyID:       transfer.CompanyID,
			BranchID:        transfer.OriginBranchID,
			ProductID:       item.ProductID,
			TransferID:      transfer.ID,
			MovementType:    "transfer_out",
			QuantityDelta:   -item.Quantity,
			StockAfter:      stockAfter,
			CreatedByUserID: valueOrZero(transfer.DispatchedByUserID),
		}); err != nil {
			return err
		}
	}

	return h.writer.Dispatch(ctx, transfer)
}
