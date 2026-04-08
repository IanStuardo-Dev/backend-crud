package transferhttp

import transferapp "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer"

func toCreateInput(request createTransferRequest, requestedByUserID int64) transferapp.CreateInput {
	items := make([]transferapp.CreateItemInput, 0, len(request.Items))
	for _, item := range request.Items {
		items = append(items, transferapp.CreateItemInput{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	return transferapp.CreateInput{
		CompanyID:           request.CompanyID,
		OriginBranchID:      request.OriginBranchID,
		DestinationBranchID: request.DestinationBranchID,
		RequestedByUserID:   requestedByUserID,
		SupervisorUserID:    request.SupervisorUserID,
		Note:                request.Note,
		Items:               items,
	}
}

func toTransferResponse(output transferapp.Output) transferResponse {
	items := make([]transferItemResponse, 0, len(output.Items))
	for _, item := range output.Items {
		items = append(items, transferItemResponse{
			ProductID:   item.ProductID,
			ProductSKU:  item.ProductSKU,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
		})
	}

	return transferResponse{
		ID:                  output.ID,
		CompanyID:           output.CompanyID,
		OriginBranchID:      output.OriginBranchID,
		DestinationBranchID: output.DestinationBranchID,
		Status:              output.Status,
		RequestedByUserID:   output.RequestedByUserID,
		SupervisorUserID:    output.SupervisorUserID,
		ApprovedByUserID:    output.ApprovedByUserID,
		DispatchedByUserID:  output.DispatchedByUserID,
		ReceivedByUserID:    output.ReceivedByUserID,
		CancelledByUserID:   output.CancelledByUserID,
		Note:                output.Note,
		Items:               items,
		CreatedAt:           output.CreatedAt,
		ApprovedAt:          output.ApprovedAt,
		DispatchedAt:        output.DispatchedAt,
		ReceivedAt:          output.ReceivedAt,
		CancelledAt:         output.CancelledAt,
	}
}

func toTransferResponses(outputs []transferapp.Output) []transferResponse {
	responses := make([]transferResponse, 0, len(outputs))
	for _, output := range outputs {
		responses = append(responses, toTransferResponse(output))
	}
	return responses
}
