package transfercommand

import (
	"context"

	transferdto "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer/dto"
	transfererrors "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer/errors"
	transferports "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer/ports"
	domaintransfer "github.com/IanStuardo-Dev/backend-crud/internal/domain/transfer"
)

type CreateHandler struct {
	txManager    transferports.TransactionManager
	writer       transferports.TransferWriter
	stockReader  transferports.InventoryTransferStockReader
	branchReader transferports.BranchReferenceReader
	userPolicies transferports.TransferUserPolicyReader
}

func NewCreateHandler(
	txManager transferports.TransactionManager,
	writer transferports.TransferWriter,
	stockReader transferports.InventoryTransferStockReader,
	branchReader transferports.BranchReferenceReader,
	userPolicies transferports.TransferUserPolicyReader,
) CreateHandler {
	return CreateHandler{
		txManager:    txManager,
		writer:       writer,
		stockReader:  stockReader,
		branchReader: branchReader,
		userPolicies: userPolicies,
	}
}

func (h CreateHandler) Handle(ctx context.Context, input transferdto.CreateInput) (transferdto.Output, error) {
	transfer := domaintransfer.Transfer{
		CompanyID:           input.CompanyID,
		OriginBranchID:      input.OriginBranchID,
		DestinationBranchID: input.DestinationBranchID,
		RequestedByUserID:   input.RequestedByUserID,
		SupervisorUserID:    input.SupervisorUserID,
		Note:                input.Note,
		Items:               toDomainItems(input.Items),
	}
	transfer.Normalize()
	if err := transfer.ValidateForCreate(); err != nil {
		return transferdto.Output{}, err
	}

	err := h.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		originExists, err := h.branchReader.BranchExists(txCtx, transfer.CompanyID, transfer.OriginBranchID)
		if err != nil {
			return err
		}
		if !originExists {
			return transfererrors.ErrInvalidReference
		}

		destinationExists, err := h.branchReader.BranchExists(txCtx, transfer.CompanyID, transfer.DestinationBranchID)
		if err != nil {
			return err
		}
		if !destinationExists {
			return transfererrors.ErrInvalidReference
		}

		requesterAllowed, err := h.userPolicies.RequesterCanAct(txCtx, transfer.CompanyID, transfer.RequestedByUserID)
		if err != nil {
			return err
		}
		if !requesterAllowed {
			return transfererrors.ErrInvalidReference
		}

		supervisorAllowed, err := h.userPolicies.SupervisorEligible(txCtx, transfer.CompanyID, transfer.SupervisorUserID)
		if err != nil {
			return err
		}
		if !supervisorAllowed {
			return transfererrors.ErrInvalidReference
		}

		productRows, err := h.stockReader.LoadForTransfer(txCtx, transfer.CompanyID, transfer.OriginBranchID, transfer.Items, false)
		if err != nil {
			return err
		}

		for index, item := range transfer.Items {
			productRow, ok := productRows[item.ProductID]
			if !ok || productRow.ProductSKU == "" {
				return transfererrors.ErrInvalidReference
			}
			transfer.Items[index].ProductSKU = productRow.ProductSKU
			transfer.Items[index].ProductName = productRow.ProductName
		}

		transfer.Status = domaintransfer.StatusPendingApproval
		return h.writer.Create(txCtx, &transfer)
	})
	if err != nil {
		return transferdto.Output{}, err
	}

	return toOutput(transfer), nil
}
