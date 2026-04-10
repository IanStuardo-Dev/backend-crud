package transfercommand

import (
	"context"
	"time"

	transferdto "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer/dto"
	transfererrors "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer/errors"
	transferports "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer/ports"
	domaintransfer "github.com/IanStuardo-Dev/backend-crud/internal/domain/transfer"
)

type TransitionHandler struct {
	txManager      transferports.TransactionManager
	reader         transferports.TransferReader
	writer         transferports.TransferWriter
	stockReader    transferports.InventoryTransferStockReader
	stockWriter    transferports.InventoryStockWriter
	movementWriter transferports.InventoryMovementWriter
}

func NewTransitionHandler(
	txManager transferports.TransactionManager,
	reader transferports.TransferReader,
	writer transferports.TransferWriter,
	stockReader transferports.InventoryTransferStockReader,
	stockWriter transferports.InventoryStockWriter,
	movementWriter transferports.InventoryMovementWriter,
) TransitionHandler {
	return TransitionHandler{
		txManager:      txManager,
		reader:         reader,
		writer:         writer,
		stockReader:    stockReader,
		stockWriter:    stockWriter,
		movementWriter: movementWriter,
	}
}

func (h TransitionHandler) Approve(ctx context.Context, input transferdto.TransitionInput) (transferdto.Output, error) {
	return h.run(ctx, input, func(transfer *domaintransfer.Transfer) error {
		return transfer.Approve(input.ActorUserID, time.Now().UTC())
	}, func(txCtx context.Context, transfer *domaintransfer.Transfer) error {
		return h.writer.Approve(txCtx, transfer)
	})
}

func (h TransitionHandler) Dispatch(ctx context.Context, input transferdto.TransitionInput) (transferdto.Output, error) {
	return h.run(ctx, input, func(transfer *domaintransfer.Transfer) error {
		return transfer.Dispatch(input.ActorUserID, time.Now().UTC())
	}, h.persistDispatch)
}

func (h TransitionHandler) Receive(ctx context.Context, input transferdto.TransitionInput) (transferdto.Output, error) {
	return h.run(ctx, input, func(transfer *domaintransfer.Transfer) error {
		return transfer.Receive(input.ActorUserID, time.Now().UTC())
	}, h.persistReceive)
}

func (h TransitionHandler) Cancel(ctx context.Context, input transferdto.TransitionInput) (transferdto.Output, error) {
	return h.run(ctx, input, func(transfer *domaintransfer.Transfer) error {
		return transfer.Cancel(input.ActorUserID, time.Now().UTC())
	}, h.persistCancel)
}

func (h TransitionHandler) run(
	ctx context.Context,
	input transferdto.TransitionInput,
	transition func(*domaintransfer.Transfer) error,
	persist func(context.Context, *domaintransfer.Transfer) error,
) (transferdto.Output, error) {
	transfer, err := h.getTransfer(ctx, input.ID)
	if err != nil {
		return transferdto.Output{}, err
	}
	if err := transition(transfer); err != nil {
		return transferdto.Output{}, mapTransitionError(err)
	}

	if err := h.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		return persist(txCtx, transfer)
	}); err != nil {
		return transferdto.Output{}, err
	}

	return toOutput(*transfer), nil
}

func (h TransitionHandler) getTransfer(ctx context.Context, id int64) (*domaintransfer.Transfer, error) {
	if id <= 0 {
		return nil, domaintransfer.ValidationError{Field: "id", Message: "id must be greater than 0"}
	}

	transfer, err := h.reader.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if transfer == nil {
		return nil, transfererrors.ErrNotFound
	}

	return transfer, nil
}
