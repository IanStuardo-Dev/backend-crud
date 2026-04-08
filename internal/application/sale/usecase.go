package saleapp

import (
	"context"

	domainsale "github.com/example/crud/internal/domain/sale"
)

type useCase struct {
	repo Repository
}

func NewUseCase(repo Repository) UseCase {
	return &useCase{repo: repo}
}

func (uc *useCase) Create(ctx context.Context, input CreateInput) (Output, error) {
	sale := domainsale.Sale{
		CompanyID:       input.CompanyID,
		BranchID:        input.BranchID,
		CreatedByUserID: input.CreatedByUserID,
		Items:           toDomainItems(input.Items),
	}
	if err := sale.ValidateForCreate(); err != nil {
		return Output{}, err
	}
	if err := uc.repo.Create(ctx, &sale); err != nil {
		return Output{}, err
	}

	return toOutput(sale), nil
}

func (uc *useCase) List(ctx context.Context) ([]Output, error) {
	sales, err := uc.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	outputs := make([]Output, 0, len(sales))
	for _, sale := range sales {
		outputs = append(outputs, toOutput(sale))
	}

	return outputs, nil
}

func (uc *useCase) GetByID(ctx context.Context, id int64) (Output, error) {
	if id <= 0 {
		return Output{}, domainsale.ValidationError{Field: "id", Message: "id must be greater than 0"}
	}

	sale, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return Output{}, err
	}
	if sale == nil {
		return Output{}, ErrNotFound
	}

	return toOutput(*sale), nil
}

func toDomainItems(inputs []CreateItemInput) []domainsale.Item {
	items := make([]domainsale.Item, 0, len(inputs))
	for _, input := range inputs {
		items = append(items, domainsale.Item{
			ProductID: input.ProductID,
			Quantity:  input.Quantity,
		})
	}

	return items
}

func toOutput(sale domainsale.Sale) Output {
	items := make([]ItemOutput, 0, len(sale.Items))
	for _, item := range sale.Items {
		items = append(items, ItemOutput{
			ProductID:      item.ProductID,
			ProductSKU:     item.ProductSKU,
			ProductName:    item.ProductName,
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
			SubtotalCents:  item.SubtotalCents,
		})
	}

	return Output{
		ID:               sale.ID,
		CompanyID:        sale.CompanyID,
		BranchID:         sale.BranchID,
		CreatedByUserID:  sale.CreatedByUserID,
		TotalAmountCents: sale.TotalAmountCents,
		Items:            items,
		CreatedAt:        sale.CreatedAt,
	}
}
