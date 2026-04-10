package transferquery

import (
	transferdto "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer/dto"
	domaintransfer "github.com/IanStuardo-Dev/backend-crud/internal/domain/transfer"
)

func toOutputs(transfers []domaintransfer.Transfer) []transferdto.Output {
	outputs := make([]transferdto.Output, 0, len(transfers))
	for _, transfer := range transfers {
		outputs = append(outputs, toOutput(transfer))
	}
	return outputs
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
