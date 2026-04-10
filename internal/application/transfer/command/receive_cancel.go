package transfercommand

import (
	"context"

	transfererrors "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer/errors"
	transferports "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer/ports"
	domaintransfer "github.com/IanStuardo-Dev/backend-crud/internal/domain/transfer"
)

func (h TransitionHandler) persistReceive(ctx context.Context, transfer *domaintransfer.Transfer) error {
	destinationRows, err := h.stockReader.LoadForTransfer(ctx, transfer.CompanyID, transfer.DestinationBranchID, transfer.Items, true)
	if err != nil {
		return err
	}

	for _, item := range transfer.Items {
		destinationRow := destinationRows[item.ProductID]
		stockAfter := destinationRow.StockOnHand + item.Quantity
		if err := h.stockWriter.PutStockOnHand(ctx, transfer.CompanyID, transfer.DestinationBranchID, item.ProductID, stockAfter); err != nil {
			return err
		}
		if err := h.movementWriter.CreateTransferMovement(ctx, transferports.MovementInput{
			CompanyID:       transfer.CompanyID,
			BranchID:        transfer.DestinationBranchID,
			ProductID:       item.ProductID,
			TransferID:      transfer.ID,
			MovementType:    "transfer_in",
			QuantityDelta:   item.Quantity,
			StockAfter:      stockAfter,
			CreatedByUserID: valueOrZero(transfer.ReceivedByUserID),
		}); err != nil {
			return err
		}
	}

	return h.writer.Receive(ctx, transfer)
}

func (h TransitionHandler) persistCancel(ctx context.Context, transfer *domaintransfer.Transfer) error {
	if transfer.DispatchedAt != nil && transfer.ReceivedAt == nil {
		originRows, err := h.stockReader.LoadForTransfer(ctx, transfer.CompanyID, transfer.OriginBranchID, transfer.Items, true)
		if err != nil {
			return err
		}

		for _, item := range transfer.Items {
			originRow, ok := originRows[item.ProductID]
			if !ok || originRow.ProductSKU == "" {
				return transfererrors.ErrInvalidReference
			}

			stockAfter := originRow.StockOnHand + item.Quantity
			if err := h.stockWriter.PutStockOnHand(ctx, transfer.CompanyID, transfer.OriginBranchID, item.ProductID, stockAfter); err != nil {
				return err
			}
			if err := h.movementWriter.CreateTransferMovement(ctx, transferports.MovementInput{
				CompanyID:       transfer.CompanyID,
				BranchID:        transfer.OriginBranchID,
				ProductID:       item.ProductID,
				TransferID:      transfer.ID,
				MovementType:    "transfer_return",
				QuantityDelta:   item.Quantity,
				StockAfter:      stockAfter,
				CreatedByUserID: valueOrZero(transfer.CancelledByUserID),
			}); err != nil {
				return err
			}
		}
	}

	return h.writer.Cancel(ctx, transfer)
}

func valueOrZero(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}
