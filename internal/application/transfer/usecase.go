package transferapp

import (
	"context"
	"errors"
	"time"

	domaintransfer "github.com/IanStuardo-Dev/backend-crud/internal/domain/transfer"
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
		SupervisorUserID:    input.SupervisorUserID,
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

func (uc *useCase) Approve(ctx context.Context, input TransitionInput) (Output, error) {
	transfer, err := uc.getTransferForTransition(ctx, input.ID)
	if err != nil {
		return Output{}, err
	}
	if err := transfer.Approve(input.ActorUserID, time.Now().UTC()); err != nil {
		return Output{}, mapTransitionError(err)
	}
	if err := uc.repo.Approve(ctx, transfer); err != nil {
		return Output{}, err
	}

	return toOutput(*transfer), nil
}

func (uc *useCase) Dispatch(ctx context.Context, input TransitionInput) (Output, error) {
	transfer, err := uc.getTransferForTransition(ctx, input.ID)
	if err != nil {
		return Output{}, err
	}
	if err := transfer.Dispatch(input.ActorUserID, time.Now().UTC()); err != nil {
		return Output{}, mapTransitionError(err)
	}
	if err := uc.repo.Dispatch(ctx, transfer); err != nil {
		return Output{}, err
	}

	return toOutput(*transfer), nil
}

func (uc *useCase) Receive(ctx context.Context, input TransitionInput) (Output, error) {
	transfer, err := uc.getTransferForTransition(ctx, input.ID)
	if err != nil {
		return Output{}, err
	}
	if err := transfer.Receive(input.ActorUserID, time.Now().UTC()); err != nil {
		return Output{}, mapTransitionError(err)
	}
	if err := uc.repo.Receive(ctx, transfer); err != nil {
		return Output{}, err
	}

	return toOutput(*transfer), nil
}

func (uc *useCase) Cancel(ctx context.Context, input TransitionInput) (Output, error) {
	transfer, err := uc.getTransferForTransition(ctx, input.ID)
	if err != nil {
		return Output{}, err
	}
	if err := transfer.Cancel(input.ActorUserID, time.Now().UTC()); err != nil {
		return Output{}, mapTransitionError(err)
	}
	if err := uc.repo.Cancel(ctx, transfer); err != nil {
		return Output{}, err
	}

	return toOutput(*transfer), nil
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

func (uc *useCase) ListByBranch(ctx context.Context, branchID int64) ([]Output, error) {
	if branchID <= 0 {
		return nil, domaintransfer.ValidationError{Field: "branch_id", Message: "branch_id must be greater than 0"}
	}

	transfers, err := uc.repo.ListByBranch(ctx, branchID)
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

func (uc *useCase) getTransferForTransition(ctx context.Context, id int64) (*domaintransfer.Transfer, error) {
	if id <= 0 {
		return nil, domaintransfer.ValidationError{Field: "id", Message: "id must be greater than 0"}
	}

	transfer, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if transfer == nil {
		return nil, ErrNotFound
	}

	return transfer, nil
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
			return errors.Join(ErrForbiddenAction, err)
		case domaintransfer.TransitionInvalidState:
			return errors.Join(ErrInvalidState, err)
		}
	}

	return err
}
