package saleapp

import (
	"context"
	"errors"
	"testing"
	"time"

	domainsale "github.com/IanStuardo-Dev/backend-crud/internal/domain/sale"
)

type stubRepository struct {
	createFn             func(context.Context, *domainsale.Sale) error
	listFn               func(context.Context) ([]domainsale.Sale, error)
	getByIDFn            func(context.Context, int64) (*domainsale.Sale, error)
	branchExistsFn       func(context.Context, int64, int64) (bool, error)
	userExistsFn         func(context.Context, int64) (bool, error)
	loadForSaleFn        func(context.Context, int64, int64, []domainsale.Item, bool) (map[int64]StockSnapshot, error)
	setStockOnHandFn     func(context.Context, int64, int64, int64, int) error
	createSaleMovementFn func(context.Context, MovementInput) error
}

func (s stubRepository) Create(ctx context.Context, sale *domainsale.Sale) error {
	if s.createFn != nil {
		return s.createFn(ctx, sale)
	}
	return nil
}

func (s stubRepository) List(ctx context.Context) ([]domainsale.Sale, error) {
	if s.listFn != nil {
		return s.listFn(ctx)
	}
	return nil, nil
}

func (s stubRepository) GetByID(ctx context.Context, id int64) (*domainsale.Sale, error) {
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

func (s stubRepository) UserExists(ctx context.Context, userID int64) (bool, error) {
	if s.userExistsFn != nil {
		return s.userExistsFn(ctx, userID)
	}
	return true, nil
}

func (s stubRepository) LoadForSale(ctx context.Context, companyID, branchID int64, items []domainsale.Item, lock bool) (map[int64]StockSnapshot, error) {
	if s.loadForSaleFn != nil {
		return s.loadForSaleFn(ctx, companyID, branchID, items, lock)
	}

	result := make(map[int64]StockSnapshot, len(items))
	for _, item := range items {
		result[item.ProductID] = StockSnapshot{
			ProductID:      item.ProductID,
			SKU:            "SKU-TEST",
			Name:           "Product",
			PriceCents:     6000,
			StockOnHand:    10,
			ReservedStock:  0,
			AvailableStock: 10,
		}
	}
	return result, nil
}

func (s stubRepository) SetStockOnHand(ctx context.Context, companyID, branchID, productID int64, stockOnHand int) error {
	if s.setStockOnHandFn != nil {
		return s.setStockOnHandFn(ctx, companyID, branchID, productID, stockOnHand)
	}
	return nil
}

func (s stubRepository) CreateSaleMovement(ctx context.Context, input MovementInput) error {
	if s.createSaleMovementFn != nil {
		return s.createSaleMovementFn(ctx, input)
	}
	return nil
}

func TestUseCaseCreateValidatesAndDelegates(t *testing.T) {
	now := time.Now().UTC()
	useCase := NewUseCase(stubRepository{
		createFn: func(_ context.Context, sale *domainsale.Sale) error {
			sale.ID = 5
			sale.CreatedAt = now
			sale.TotalAmountCents = 12000
			sale.Items = []domainsale.Item{{
				ProductID:      9,
				ProductSKU:     "SKU-009",
				ProductName:    "Monitor",
				Quantity:       2,
				UnitPriceCents: 6000,
				SubtotalCents:  12000,
			}}
			return nil
		},
	})

	output, err := useCase.Create(context.Background(), CreateInput{
		CompanyID:       1,
		BranchID:        1,
		CreatedByUserID: 3,
		Items: []CreateItemInput{{
			ProductID: 9,
			Quantity:  2,
		}},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if output.ID != 5 || output.TotalAmountCents != 12000 {
		t.Fatalf("unexpected output: %#v", output)
	}
	if len(output.Items) != 1 || output.Items[0].ProductSKU != "SKU-009" {
		t.Fatalf("unexpected items: %#v", output.Items)
	}
}

func TestUseCaseCreateRejectsDuplicateProducts(t *testing.T) {
	called := false
	useCase := NewUseCase(stubRepository{
		createFn: func(_ context.Context, sale *domainsale.Sale) error {
			called = true
			return nil
		},
	})

	_, err := useCase.Create(context.Background(), CreateInput{
		CompanyID:       1,
		BranchID:        1,
		CreatedByUserID: 2,
		Items: []CreateItemInput{
			{ProductID: 9, Quantity: 1},
			{ProductID: 9, Quantity: 2},
		},
	})

	var validationErr domainsale.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if called {
		t.Fatal("expected repository Create not to be called")
	}
}

func TestUseCaseGetByIDReturnsNotFound(t *testing.T) {
	useCase := NewUseCase(stubRepository{
		getByIDFn: func(context.Context, int64) (*domainsale.Sale, error) {
			return nil, nil
		},
	})

	_, err := useCase.GetByID(context.Background(), 10)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUseCaseListDelegates(t *testing.T) {
	now := time.Now().UTC()
	useCase := NewUseCase(stubRepository{
		listFn: func(context.Context) ([]domainsale.Sale, error) {
			return []domainsale.Sale{{
				ID:               1,
				CompanyID:        1,
				BranchID:         1,
				CreatedByUserID:  3,
				TotalAmountCents: 9000,
				CreatedAt:        now,
			}}, nil
		},
	})

	sales, err := useCase.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(sales) != 1 || sales[0].TotalAmountCents != 9000 {
		t.Fatalf("unexpected sales: %#v", sales)
	}
}
