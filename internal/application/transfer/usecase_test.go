package transferapp

import (
	"context"
	"errors"
	"testing"
	"time"

	domaintransfer "github.com/IanStuardo-Dev/backend-crud/internal/domain/transfer"
)

type stubRepository struct {
	createFn             func(context.Context, *domaintransfer.Transfer) error
	approveFn            func(context.Context, *domaintransfer.Transfer) error
	dispatchFn           func(context.Context, *domaintransfer.Transfer) error
	receiveFn            func(context.Context, *domaintransfer.Transfer) error
	cancelFn             func(context.Context, *domaintransfer.Transfer) error
	listFn               func(context.Context) ([]domaintransfer.Transfer, error)
	getByIDFn            func(context.Context, int64) (*domaintransfer.Transfer, error)
	branchExistsFn       func(context.Context, int64, int64) (bool, error)
	requesterCanActFn    func(context.Context, int64, int64) (bool, error)
	supervisorEligibleFn func(context.Context, int64, int64) (bool, error)
	loadForTransferFn    func(context.Context, int64, int64, []domaintransfer.Item, bool) (map[int64]StockSnapshot, error)
	putStockOnHandFn     func(context.Context, int64, int64, int64, int) error
	createMovementFn     func(context.Context, MovementInput) error
}

func (s stubRepository) Create(ctx context.Context, transfer *domaintransfer.Transfer) error {
	if s.createFn != nil {
		return s.createFn(ctx, transfer)
	}
	return nil
}

func (s stubRepository) Approve(ctx context.Context, transfer *domaintransfer.Transfer) error {
	if s.approveFn != nil {
		return s.approveFn(ctx, transfer)
	}
	return nil
}

func (s stubRepository) Dispatch(ctx context.Context, transfer *domaintransfer.Transfer) error {
	if s.dispatchFn != nil {
		return s.dispatchFn(ctx, transfer)
	}
	return nil
}

func (s stubRepository) Receive(ctx context.Context, transfer *domaintransfer.Transfer) error {
	if s.receiveFn != nil {
		return s.receiveFn(ctx, transfer)
	}
	return nil
}

func (s stubRepository) Cancel(ctx context.Context, transfer *domaintransfer.Transfer) error {
	if s.cancelFn != nil {
		return s.cancelFn(ctx, transfer)
	}
	return nil
}

func (s stubRepository) List(ctx context.Context) ([]domaintransfer.Transfer, error) {
	if s.listFn != nil {
		return s.listFn(ctx)
	}
	return nil, nil
}

func (s stubRepository) ListByBranch(ctx context.Context, branchID int64) ([]domaintransfer.Transfer, error) {
	if s.listFn != nil {
		return s.listFn(ctx)
	}
	return nil, nil
}

func (s stubRepository) GetByID(ctx context.Context, id int64) (*domaintransfer.Transfer, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (s stubRepository) BranchExists(ctx context.Context, companyID, branchID int64) (bool, error) {
	if s.branchExistsFn != nil {
		return s.branchExistsFn(ctx, companyID, branchID)
	}
	return true, nil
}

func (s stubRepository) RequesterCanAct(ctx context.Context, companyID, userID int64) (bool, error) {
	if s.requesterCanActFn != nil {
		return s.requesterCanActFn(ctx, companyID, userID)
	}
	return true, nil
}

func (s stubRepository) SupervisorEligible(ctx context.Context, companyID, userID int64) (bool, error) {
	if s.supervisorEligibleFn != nil {
		return s.supervisorEligibleFn(ctx, companyID, userID)
	}
	return true, nil
}

func (s stubRepository) LoadForTransfer(ctx context.Context, companyID, branchID int64, items []domaintransfer.Item, lock bool) (map[int64]StockSnapshot, error) {
	if s.loadForTransferFn != nil {
		return s.loadForTransferFn(ctx, companyID, branchID, items, lock)
	}
	result := make(map[int64]StockSnapshot, len(items))
	for _, item := range items {
		result[item.ProductID] = StockSnapshot{
			ProductID:      item.ProductID,
			ProductSKU:     "SKU-TEST",
			ProductName:    "Product",
			StockOnHand:    10,
			ReservedStock:  0,
			AvailableStock: 10,
		}
	}
	return result, nil
}

func (s stubRepository) PutStockOnHand(ctx context.Context, companyID, branchID, productID int64, stockOnHand int) error {
	if s.putStockOnHandFn != nil {
		return s.putStockOnHandFn(ctx, companyID, branchID, productID, stockOnHand)
	}
	return nil
}

func (s stubRepository) CreateTransferMovement(ctx context.Context, input MovementInput) error {
	if s.createMovementFn != nil {
		return s.createMovementFn(ctx, input)
	}
	return nil
}

func TestUseCaseCreateRequiresSupervisor(t *testing.T) {
	called := false
	useCase := NewUseCase(stubRepository{
		createFn: func(_ context.Context, _ *domaintransfer.Transfer) error {
			called = true
			return nil
		},
	})

	_, err := useCase.Create(context.Background(), CreateInput{
		CompanyID:           1,
		OriginBranchID:      1,
		DestinationBranchID: 2,
		RequestedByUserID:   7,
		SupervisorUserID:    7,
		Items: []CreateItemInput{{
			ProductID: 10,
			Quantity:  1,
		}},
	})

	var validationErr domaintransfer.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if called {
		t.Fatal("expected repository Create not to be called")
	}
}

func TestUseCaseApproveMapsForbiddenSupervisorAction(t *testing.T) {
	useCase := NewUseCase(stubRepository{
		getByIDFn: func(context.Context, int64) (*domaintransfer.Transfer, error) {
			return &domaintransfer.Transfer{
				ID:                3,
				CompanyID:         1,
				Status:            domaintransfer.StatusPendingApproval,
				RequestedByUserID: 7,
				SupervisorUserID:  9,
			}, nil
		},
	})

	_, err := useCase.Approve(context.Background(), TransitionInput{ID: 3, ActorUserID: 8})
	if !errors.Is(err, ErrForbiddenAction) {
		t.Fatalf("expected ErrForbiddenAction, got %v", err)
	}
}

func TestUseCaseReceiveDelegatesTransition(t *testing.T) {
	now := time.Now().UTC()
	useCase := NewUseCase(stubRepository{
		getByIDFn: func(context.Context, int64) (*domaintransfer.Transfer, error) {
			return &domaintransfer.Transfer{
				ID:                  5,
				CompanyID:           1,
				OriginBranchID:      1,
				DestinationBranchID: 2,
				Status:              domaintransfer.StatusInTransit,
				RequestedByUserID:   7,
				SupervisorUserID:    9,
				Items: []domaintransfer.Item{{
					ProductID: 10,
					Quantity:  2,
				}},
				DispatchedAt: &now,
			}, nil
		},
		receiveFn: func(_ context.Context, transfer *domaintransfer.Transfer) error {
			if transfer.Status != domaintransfer.StatusReceived {
				t.Fatalf("expected transfer status received, got %s", transfer.Status)
			}
			return nil
		},
	})

	output, err := useCase.Receive(context.Background(), TransitionInput{ID: 5, ActorUserID: 11})
	if err != nil {
		t.Fatalf("Receive() error = %v", err)
	}
	if output.Status != domaintransfer.StatusReceived || output.ReceivedByUserID == nil || *output.ReceivedByUserID != 11 {
		t.Fatalf("unexpected output: %#v", output)
	}
}

func TestUseCaseListByBranchValidatesBranchID(t *testing.T) {
	useCase := NewUseCase(stubRepository{})

	_, err := useCase.ListByBranch(context.Background(), 0)
	var validationErr domaintransfer.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if validationErr.Field != "branch_id" {
		t.Fatalf("expected branch_id validation, got %#v", validationErr)
	}
}
