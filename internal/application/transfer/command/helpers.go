package transfercommand

import (
	"errors"

	transferdto "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer/dto"
	transfererrors "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer/errors"
	domaintransfer "github.com/IanStuardo-Dev/backend-crud/internal/domain/transfer"
)

func toDomainItems(inputs []transferdto.CreateItemInput) []domaintransfer.Item {
	items := make([]domaintransfer.Item, 0, len(inputs))
	for _, input := range inputs {
		items = append(items, domaintransfer.Item{
			ProductID: input.ProductID,
			Quantity:  input.Quantity,
		})
	}

	return items
}

func toOutput(transfer domaintransfer.Transfer) transferdto.Output {
	items := make([]transferdto.ItemOutput, 0, len(transfer.Items))
	for _, item := range transfer.Items {
		items = append(items, transferdto.ItemOutput{
			ProductID:   item.ProductID,
			ProductSKU:  item.ProductSKU,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
		})
	}

	return transferdto.Output{
		ID:                  transfer.ID,
		CompanyID:           transfer.CompanyID,
		OriginBranchID:      transfer.OriginBranchID,
		DestinationBranchID: transfer.DestinationBranchID,
		Status:              transfer.Status,
		RequestedByUserID:   transfer.RequestedByUserID,
		SupervisorUserID:    transfer.SupervisorUserID,
		ApprovedByUserID:    transfer.ApprovedByUserID,
		DispatchedByUserID:  transfer.DispatchedByUserID,
		ReceivedByUserID:    transfer.ReceivedByUserID,
		CancelledByUserID:   transfer.CancelledByUserID,
		Note:                transfer.Note,
		Items:               items,
		CreatedAt:           transfer.CreatedAt,
		ApprovedAt:          transfer.ApprovedAt,
		DispatchedAt:        transfer.DispatchedAt,
		ReceivedAt:          transfer.ReceivedAt,
		CancelledAt:         transfer.CancelledAt,
	}
}

func mapTransitionError(err error) error {
	var transitionErr domaintransfer.TransitionError
	if errors.As(err, &transitionErr) {
		switch transitionErr.Kind {
		case domaintransfer.TransitionForbidden:
			return errors.Join(transfererrors.ErrForbiddenAction, err)
		case domaintransfer.TransitionInvalidState:
			return errors.Join(transfererrors.ErrInvalidState, err)
		}
	}

	return err
}
