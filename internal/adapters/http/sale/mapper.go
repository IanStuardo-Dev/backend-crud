package salehttp

import saleapp "github.com/IanStuardo-Dev/backend-crud/internal/application/sale"

func toCreateInput(request createSaleRequest, userID int64) saleapp.CreateInput {
	items := make([]saleapp.CreateItemInput, 0, len(request.Items))
	for _, item := range request.Items {
		items = append(items, saleapp.CreateItemInput{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	return saleapp.CreateInput{
		CompanyID:       request.CompanyID,
		BranchID:        request.BranchID,
		CreatedByUserID: userID,
		Items:           items,
	}
}

func toSaleResponse(output saleapp.Output) saleResponse {
	items := make([]saleItemResponse, 0, len(output.Items))
	for _, item := range output.Items {
		items = append(items, saleItemResponse{
			ProductID:      item.ProductID,
			ProductSKU:     item.ProductSKU,
			ProductName:    item.ProductName,
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
			SubtotalCents:  item.SubtotalCents,
		})
	}

	return saleResponse{
		ID:               output.ID,
		CompanyID:        output.CompanyID,
		BranchID:         output.BranchID,
		CreatedByUserID:  output.CreatedByUserID,
		TotalAmountCents: output.TotalAmountCents,
		Items:            items,
		CreatedAt:        output.CreatedAt,
	}
}

func toSaleResponses(outputs []saleapp.Output) []saleResponse {
	responses := make([]saleResponse, 0, len(outputs))
	for _, output := range outputs {
		responses = append(responses, toSaleResponse(output))
	}
	return responses
}
