package inventoryhttp

import inventoryapp "github.com/example/crud/internal/application/inventory"

func toListByBranchInput(request listByBranchRequest) inventoryapp.ListByBranchInput {
	return inventoryapp.ListByBranchInput{
		CompanyID: request.CompanyID,
		BranchID:  request.BranchID,
	}
}

func toSuggestSourcesInput(request suggestSourcesRequest) inventoryapp.SuggestSourcesInput {
	return inventoryapp.SuggestSourcesInput{
		CompanyID:           request.CompanyID,
		DestinationBranchID: request.DestinationBranchID,
		ProductID:           request.ProductID,
		Quantity:            request.Quantity,
	}
}

func toItemResponses(outputs []inventoryapp.ItemOutput) []itemResponse {
	responses := make([]itemResponse, 0, len(outputs))
	for _, output := range outputs {
		responses = append(responses, itemResponse{
			CompanyID:      output.CompanyID,
			BranchID:       output.BranchID,
			ProductID:      output.ProductID,
			ProductSKU:     output.ProductSKU,
			ProductName:    output.ProductName,
			Category:       output.Category,
			Brand:          output.Brand,
			StockOnHand:    output.StockOnHand,
			ReservedStock:  output.ReservedStock,
			AvailableStock: output.AvailableStock,
		})
	}

	return responses
}

func toSourceCandidateResponses(outputs []inventoryapp.SourceCandidateOutput) []sourceCandidateResponse {
	responses := make([]sourceCandidateResponse, 0, len(outputs))
	for _, output := range outputs {
		responses = append(responses, sourceCandidateResponse{
			CompanyID:          output.CompanyID,
			BranchID:           output.BranchID,
			BranchCode:         output.BranchCode,
			BranchName:         output.BranchName,
			City:               output.City,
			Region:             output.Region,
			Latitude:           output.Latitude,
			Longitude:          output.Longitude,
			ProductID:          output.ProductID,
			AvailableStock:     output.AvailableStock,
			DistanceKilometers: output.DistanceKilometers,
		})
	}

	return responses
}
