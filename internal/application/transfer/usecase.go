package transferapp

import (
	"context"

	domaintransfer "github.com/example/crud/internal/domain/transfer"
)

type useCase struct {
	repo Repository
}

func NewUseCase(repo Repository) UseCase {
	return &useCase{repo: repo}
}

func (uc *useCase) Create(ctx context.Context, input CreateInput) (Output, error) {
	transfer := domaintransfer.Transfer{
		CompanyID:           input.CompanyID,
		OriginBranchID:      input.OriginBranchID,
		DestinationBranchID: input.DestinationBranchID,
		RequestedByUserID:   input.RequestedByUserID,
		CompletedByUserID:   input.RequestedByUserID,
		Note:                input.Note,
		Items:               toDomainItems(input.Items),
	}
	transfer.Normalize()
	if err := transfer.ValidateForCreate(); err != nil {
		return Output{}, err
	}
	if err := uc.repo.Create(ctx, &transfer); err != nil {
		return Output{}, err
	}

	return toOutput(transfer), nil
}

func (uc *useCase) List(ctx context.Context) ([]Output, error) {
	transfers, err := uc.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	outputs := make([]Output, 0, len(transfers))
	for _, transfer := range transfers {
		outputs = append(outputs, toOutput(transfer))
	}

	return outputs, nil
}

func (uc *useCase) GetByID(ctx context.Context, id int64) (Output, error) {
	if id <= 0 {
		return Output{}, domaintransfer.ValidationError{Field: "id", Message: "id must be greater than 0"}
	}

	transfer, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return Output{}, err
	}
	if transfer == nil {
		return Output{}, ErrNotFound
	}

	return toOutput(*transfer), nil
}

func toDomainItems(inputs []CreateItemInput) []domaintransfer.Item {
	items := make([]domaintransfer.Item, 0, len(inputs))
	for _, input := range inputs {
		items = append(items, domaintransfer.Item{
			ProductID: input.ProductID,
			Quantity:  input.Quantity,
		})
	}

	return items
}

func toOutput(transfer domaintransfer.Transfer) Output {
	items := make([]ItemOutput, 0, len(transfer.Items))
	for _, item := range transfer.Items {
		items = append(items, ItemOutput{
			ProductID:   item.ProductID,
			ProductSKU:  item.ProductSKU,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
		})
	}

	return Output{
		ID:                  transfer.ID,
		CompanyID:           transfer.CompanyID,
		OriginBranchID:      transfer.OriginBranchID,
		DestinationBranchID: transfer.DestinationBranchID,
		Status:              transfer.Status,
		RequestedByUserID:   transfer.RequestedByUserID,
		CompletedByUserID:   transfer.CompletedByUserID,
		Note:                transfer.Note,
		Items:               items,
		CreatedAt:           transfer.CreatedAt,
		CompletedAt:         transfer.CompletedAt,
	}
}
