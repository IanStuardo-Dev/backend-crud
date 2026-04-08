package inventoryapp

import (
	"context"

	domaininventory "github.com/example/crud/internal/domain/inventory"
)

type useCase struct {
	repo Repository
}

func NewUseCase(repo Repository) UseCase {
	return &useCase{repo: repo}
}

func (uc *useCase) ListByBranch(ctx context.Context, input ListByBranchInput) ([]ItemOutput, error) {
	if input.CompanyID <= 0 {
		return nil, domaininventory.ValidationError{Field: "company_id", Message: "company_id must be greater than 0"}
	}
	if input.BranchID <= 0 {
		return nil, domaininventory.ValidationError{Field: "branch_id", Message: "branch_id must be greater than 0"}
	}

	items, err := uc.repo.ListByBranch(ctx, input.CompanyID, input.BranchID)
	if err != nil {
		return nil, err
	}

	outputs := make([]ItemOutput, 0, len(items))
	for _, item := range items {
		outputs = append(outputs, ItemOutput{
			CompanyID:      item.CompanyID,
			BranchID:       item.BranchID,
			ProductID:      item.ProductID,
			ProductSKU:     item.ProductSKU,
			ProductName:    item.ProductName,
			Category:       item.Category,
			Brand:          item.Brand,
			StockOnHand:    item.StockOnHand,
			ReservedStock:  item.ReservedStock,
			AvailableStock: item.AvailableStock,
		})
	}

	return outputs, nil
}

func (uc *useCase) SuggestSources(ctx context.Context, input SuggestSourcesInput) ([]SourceCandidateOutput, error) {
	if input.CompanyID <= 0 {
		return nil, domaininventory.ValidationError{Field: "company_id", Message: "company_id must be greater than 0"}
	}
	if input.DestinationBranchID <= 0 {
		return nil, domaininventory.ValidationError{Field: "destination_branch_id", Message: "destination_branch_id must be greater than 0"}
	}
	if input.ProductID <= 0 {
		return nil, domaininventory.ValidationError{Field: "product_id", Message: "product_id must be greater than 0"}
	}
	if input.Quantity <= 0 {
		return nil, domaininventory.ValidationError{Field: "quantity", Message: "quantity must be greater than 0"}
	}

	candidates, err := uc.repo.SuggestSources(ctx, input.CompanyID, input.DestinationBranchID, input.ProductID, input.Quantity)
	if err != nil {
		return nil, err
	}

	outputs := make([]SourceCandidateOutput, 0, len(candidates))
	for _, candidate := range candidates {
		outputs = append(outputs, SourceCandidateOutput{
			CompanyID:          candidate.CompanyID,
			BranchID:           candidate.BranchID,
			BranchCode:         candidate.BranchCode,
			BranchName:         candidate.BranchName,
			City:               candidate.City,
			Region:             candidate.Region,
			Latitude:           candidate.Latitude,
			Longitude:          candidate.Longitude,
			ProductID:          candidate.ProductID,
			AvailableStock:     candidate.AvailableStock,
			DistanceKilometers: candidate.DistanceKilometers,
		})
	}

	return outputs, nil
}
