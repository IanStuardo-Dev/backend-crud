package producthttp

import (
	"strconv"

	productapp "github.com/IanStuardo-Dev/backend-crud/internal/application/product"
)

func toCreateInput(request createProductRequest) productapp.CreateInput {
	return productapp.CreateInput{
		CompanyID:   request.CompanyID,
		BranchID:    request.BranchID,
		SKU:         request.SKU,
		Name:        request.Name,
		Description: request.Description,
		Category:    request.Category,
		Brand:       request.Brand,
		PriceCents:  request.PriceCents,
		Currency:    request.Currency,
		Stock:       request.Stock,
		Embedding:   cloneEmbedding(request.Embedding),
	}
}

func toUpdateInput(id int64, request updateProductRequest) productapp.UpdateInput {
	return productapp.UpdateInput{
		ID:          id,
		CompanyID:   request.CompanyID,
		BranchID:    request.BranchID,
		SKU:         request.SKU,
		Name:        request.Name,
		Description: request.Description,
		Category:    request.Category,
		Brand:       request.Brand,
		PriceCents:  request.PriceCents,
		Currency:    request.Currency,
		Stock:       request.Stock,
		Embedding:   cloneEmbedding(request.Embedding),
	}
}

func toProductResponse(output productapp.Output) productResponse {
	return productResponse{
		ID:          output.ID,
		CompanyID:   output.CompanyID,
		BranchID:    output.BranchID,
		SKU:         output.SKU,
		Name:        output.Name,
		Description: output.Description,
		Category:    output.Category,
		Brand:       output.Brand,
		PriceCents:  output.PriceCents,
		Currency:    output.Currency,
		Stock:       output.Stock,
		Embedding:   cloneEmbedding(output.Embedding),
		CreatedAt:   output.CreatedAt,
		UpdatedAt:   output.UpdatedAt,
	}
}

func toProductResponses(outputs []productapp.Output) []productResponse {
	responses := make([]productResponse, 0, len(outputs))
	for _, output := range outputs {
		responses = append(responses, toProductResponse(output))
	}

	return responses
}

func toFindNeighborsInput(id int64, limitRaw, minSimilarityRaw string) (productapp.FindNeighborsInput, error) {
	input := productapp.FindNeighborsInput{ProductID: id}

	if limitRaw != "" {
		limit, err := strconv.Atoi(limitRaw)
		if err != nil {
			return productapp.FindNeighborsInput{}, err
		}
		input.Limit = limit
	}
	if minSimilarityRaw != "" {
		minSimilarity, err := strconv.ParseFloat(minSimilarityRaw, 64)
		if err != nil {
			return productapp.FindNeighborsInput{}, err
		}
		input.MinSimilarity = minSimilarity
	}

	return input, nil
}

func toNeighborResponses(outputs []productapp.NeighborOutput) []neighborResponse {
	responses := make([]neighborResponse, 0, len(outputs))
	for _, output := range outputs {
		responses = append(responses, neighborResponse{
			ProductID:            output.ProductID,
			SKU:                  output.SKU,
			Name:                 output.Name,
			Description:          output.Description,
			Category:             output.Category,
			Brand:                output.Brand,
			PriceCents:           output.PriceCents,
			Currency:             output.Currency,
			SimilarityPercentage: output.SimilarityPercentage,
			Distance:             output.Distance,
		})
	}

	return responses
}

func toNeighborFeedbackInput(sourceProductID, suggestedProductID, userID int64, request neighborFeedbackRequest) productapp.RecordNeighborFeedbackInput {
	return productapp.RecordNeighborFeedbackInput{
		SourceProductID:    sourceProductID,
		SuggestedProductID: suggestedProductID,
		BranchID:           request.BranchID,
		UserID:             userID,
		Action:             request.Action,
		Note:               request.Note,
	}
}

func toNeighborFeedbackResponse(output productapp.NeighborFeedbackOutput) neighborFeedbackResponse {
	return neighborFeedbackResponse{
		SourceProductID:    output.SourceProductID,
		SuggestedProductID: output.SuggestedProductID,
		BranchID:           output.BranchID,
		UserID:             output.UserID,
		Action:             output.Action,
		Note:               output.Note,
		CreatedAt:          output.CreatedAt,
		UpdatedAt:          output.UpdatedAt,
	}
}

func cloneEmbedding(embedding []float32) []float32 {
	if len(embedding) == 0 {
		return nil
	}

	cloned := make([]float32, len(embedding))
	copy(cloned, embedding)
	return cloned
}
