package userapp

import (
	"context"
	"errors"
	"testing"

	domainuser "github.com/IanStuardo-Dev/backend-crud/internal/domain/user"
)

type stubRepository struct {
	createFn     func(context.Context, *domainuser.User) error
	listFn       func(context.Context) ([]domainuser.User, error)
	getByIDFn    func(context.Context, int64) (*domainuser.User, error)
	getByEmailFn func(context.Context, string) (*domainuser.User, error)
	updateFn     func(context.Context, *domainuser.User) error
	deleteFn     func(context.Context, int64) error
}

func (s stubRepository) Create(ctx context.Context, user *domainuser.User) error {
	if s.createFn != nil {
		return s.createFn(ctx, user)
	}

	return nil
}

func (s stubRepository) List(ctx context.Context) ([]domainuser.User, error) {
	if s.listFn != nil {
		return s.listFn(ctx)
	}

	return nil, nil
}

func (s stubRepository) GetByID(ctx context.Context, id int64) (*domainuser.User, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(ctx, id)
	}

	return nil, nil
}

func (s stubRepository) GetByEmail(ctx context.Context, email string) (*domainuser.User, error) {
	if s.getByEmailFn != nil {
		return s.getByEmailFn(ctx, email)
	}

	return nil, nil
}

func (s stubRepository) Update(ctx context.Context, user *domainuser.User) error {
	if s.updateFn != nil {
		return s.updateFn(ctx, user)
	}

	return nil
}

func (s stubRepository) Delete(ctx context.Context, id int64) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, id)
	}

	return nil
}

type stubHasher struct {
	hashFn func(string) (string, error)
}

func (s stubHasher) Hash(password string) (string, error) {
	if s.hashFn != nil {
		return s.hashFn(password)
	}

	return "hashed:" + password, nil
}

func TestUseCaseCreateNormalizesAndValidates(t *testing.T) {
	var createdUser domainuser.User
	useCase := NewUseCase(stubRepository{
		createFn: func(_ context.Context, user *domainuser.User) error {
			createdUser = *user
			user.ID = 7
			return nil
		},
	}, stubHasher{})

	output, err := useCase.Create(context.Background(), CreateInput{
		CompanyID:       int64Pointer(1),
		Name:            "  Alice Doe  ",
		Email:           "ALICE@EXAMPLE.COM ",
		Role:            domainuser.RoleCompanyAdmin,
		IsActive:        true,
		DefaultBranchID: int64Pointer(1),
		Password:        "Password123",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if createdUser.Name != "Alice Doe" {
		t.Fatalf("expected normalized name, got %q", createdUser.Name)
	}
	if createdUser.Email != "alice@example.com" {
		t.Fatalf("expected normalized email, got %q", createdUser.Email)
	}
	if createdUser.PasswordHash != "hashed:Password123" {
		t.Fatalf("expected password hash to be set, got %q", createdUser.PasswordHash)
	}
	if output.ID != 7 {
		t.Fatalf("expected generated ID to be returned, got %d", output.ID)
	}
}

func TestUseCaseCreateRejectsInvalidUser(t *testing.T) {
	called := false
	useCase := NewUseCase(stubRepository{
		createFn: func(_ context.Context, user *domainuser.User) error {
			called = true
			return nil
		},
	}, stubHasher{})

	_, err := useCase.Create(context.Background(), CreateInput{
		CompanyID: int64Pointer(1),
		Name:      "A",
		Email:     "bad-email",
		Role:      domainuser.RoleCompanyAdmin,
		IsActive:  true,
		Password:  "Password123",
	})

	var validationErr domainuser.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if called {
		t.Fatal("expected repository Create not to be called on invalid input")
	}
}

func TestUseCaseGetByIDReturnsNotFound(t *testing.T) {
	useCase := NewUseCase(stubRepository{
		getByIDFn: func(context.Context, int64) (*domainuser.User, error) {
			return nil, nil
		},
	}, stubHasher{})

	_, err := useCase.GetByID(context.Background(), 10)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUseCaseUpdateValidatesIDAndNormalizes(t *testing.T) {
	var updatedUser domainuser.User
	useCase := NewUseCase(stubRepository{
		updateFn: func(_ context.Context, user *domainuser.User) error {
			updatedUser = *user
			return nil
		},
	}, stubHasher{})

	output, err := useCase.Update(context.Background(), UpdateInput{
		ID:              3,
		CompanyID:       int64Pointer(1),
		Name:            "  Bob  ",
		Email:           "BOB@EXAMPLE.COM ",
		Role:            domainuser.RoleInventoryManager,
		IsActive:        true,
		DefaultBranchID: int64Pointer(1),
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if updatedUser.ID != 3 {
		t.Fatalf("expected id 3, got %d", updatedUser.ID)
	}
	if updatedUser.Name != "Bob" {
		t.Fatalf("expected normalized name, got %q", updatedUser.Name)
	}
	if updatedUser.Email != "bob@example.com" {
		t.Fatalf("expected normalized email, got %q", updatedUser.Email)
	}
	if output.Email != "bob@example.com" {
		t.Fatalf("expected output email bob@example.com, got %q", output.Email)
	}
}

func TestUseCaseUpdateRejectsInvalidID(t *testing.T) {
	called := false
	useCase := NewUseCase(stubRepository{
		updateFn: func(_ context.Context, user *domainuser.User) error {
			called = true
			return nil
		},
	}, stubHasher{})

	_, err := useCase.Update(context.Background(), UpdateInput{
		ID:        0,
		CompanyID: int64Pointer(1),
		Name:      "Alice",
		Email:     "alice@example.com",
		Role:      domainuser.RoleCompanyAdmin,
		IsActive:  true,
	})

	var validationErr domainuser.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if validationErr.Field != "id" {
		t.Fatalf("expected id field validation, got %q", validationErr.Field)
	}
	if called {
		t.Fatal("expected repository Update not to be called on invalid id")
	}
}

func TestUseCaseDeleteRejectsInvalidID(t *testing.T) {
	called := false
	useCase := NewUseCase(stubRepository{
		deleteFn: func(_ context.Context, id int64) error {
			called = true
			return nil
		},
	}, stubHasher{})

	err := useCase.Delete(context.Background(), 0)

	var validationErr domainuser.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if validationErr.Field != "id" {
		t.Fatalf("expected id field validation, got %q", validationErr.Field)
	}
	if called {
		t.Fatal("expected repository Delete not to be called on invalid id")
	}
}

func TestUseCaseDeletePropagatesNotFound(t *testing.T) {
	useCase := NewUseCase(stubRepository{
		deleteFn: func(_ context.Context, id int64) error {
			return ErrNotFound
		},
	}, stubHasher{})

	err := useCase.Delete(context.Background(), 10)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUseCaseReadOperationsDelegate(t *testing.T) {
	expectedUsers := []domainuser.User{{ID: 1, CompanyID: int64Pointer(1), Name: "Alice", Email: "alice@example.com", Role: domainuser.RoleCompanyAdmin, IsActive: true}}
	expectedUser := &domainuser.User{ID: 2, CompanyID: int64Pointer(1), Name: "Bob", Email: "bob@example.com", Role: domainuser.RoleSalesUser, IsActive: true}

	useCase := NewUseCase(stubRepository{
		listFn: func(context.Context) ([]domainuser.User, error) {
			return expectedUsers, nil
		},
		getByIDFn: func(_ context.Context, id int64) (*domainuser.User, error) {
			if id != 2 {
				t.Fatalf("expected id 2, got %d", id)
			}
			return expectedUser, nil
		},
	}, stubHasher{})

	users, err := useCase.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(users) != 1 || users[0].Email != expectedUsers[0].Email {
		t.Fatalf("unexpected users: %#v", users)
	}

	user, err := useCase.GetByID(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if user.Email != expectedUser.Email {
		t.Fatalf("unexpected user: %#v", user)
	}
}

func int64Pointer(value int64) *int64 {
	return &value
}
