package authapp

import (
	"context"
	"errors"
	"testing"
	"time"

	domainuser "github.com/example/crud/internal/domain/user"
)

type stubUserRepository struct {
	getByEmailFn func(context.Context, string) (*domainuser.User, error)
}

func (s stubUserRepository) GetByEmail(ctx context.Context, email string) (*domainuser.User, error) {
	if s.getByEmailFn != nil {
		return s.getByEmailFn(ctx, email)
	}

	return nil, nil
}

type stubPasswordVerifier struct {
	compareFn func(string, string) error
}

func (s stubPasswordVerifier) Compare(hash, password string) error {
	if s.compareFn != nil {
		return s.compareFn(hash, password)
	}

	return nil
}

type stubTokenManager struct {
	generateFn func(AuthenticatedUser) (IssuedToken, error)
	verifyFn   func(string) (AuthenticatedUser, error)
}

func (s stubTokenManager) Generate(user AuthenticatedUser) (IssuedToken, error) {
	if s.generateFn != nil {
		return s.generateFn(user)
	}

	return IssuedToken{}, nil
}

func (s stubTokenManager) Verify(token string) (AuthenticatedUser, error) {
	if s.verifyFn != nil {
		return s.verifyFn(token)
	}

	return AuthenticatedUser{}, nil
}

func TestUseCaseLoginReturnsToken(t *testing.T) {
	expiresAt := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)
	useCase := NewUseCase(
		stubUserRepository{
			getByEmailFn: func(_ context.Context, email string) (*domainuser.User, error) {
				if email != "alice@example.com" {
					t.Fatalf("expected normalized email, got %q", email)
				}
				return &domainuser.User{
					ID:              7,
					CompanyID:       int64Pointer(1),
					Name:            "Alice",
					Email:           email,
					Role:            domainuser.RoleCompanyAdmin,
					IsActive:        true,
					DefaultBranchID: int64Pointer(1),
					PasswordHash:    "hash",
				}, nil
			},
		},
		stubPasswordVerifier{
			compareFn: func(hash, password string) error {
				if hash != "hash" || password != "Password123" {
					t.Fatalf("unexpected credentials hash=%q password=%q", hash, password)
				}
				return nil
			},
		},
		stubTokenManager{
			generateFn: func(user AuthenticatedUser) (IssuedToken, error) {
				if user.ID != 7 {
					t.Fatalf("expected user id 7, got %d", user.ID)
				}
				return IssuedToken{
					AccessToken: "token-123",
					ExpiresAt:   expiresAt,
				}, nil
			},
		},
	)

	output, err := useCase.Login(context.Background(), LoginInput{
		Email:    " ALICE@EXAMPLE.COM ",
		Password: "Password123",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if output.AccessToken != "token-123" {
		t.Fatalf("expected token token-123, got %q", output.AccessToken)
	}
	if output.ExpiresAt != expiresAt {
		t.Fatalf("expected expiresAt %v, got %v", expiresAt, output.ExpiresAt)
	}
	if output.User.Email != "alice@example.com" {
		t.Fatalf("expected normalized email in output, got %q", output.User.Email)
	}
}

func TestUseCaseLoginRejectsInvalidCredentials(t *testing.T) {
	useCase := NewUseCase(
		stubUserRepository{
			getByEmailFn: func(context.Context, string) (*domainuser.User, error) {
				return &domainuser.User{
					ID:              7,
					CompanyID:       int64Pointer(1),
					Name:            "Alice",
					Email:           "alice@example.com",
					Role:            domainuser.RoleCompanyAdmin,
					IsActive:        true,
					DefaultBranchID: int64Pointer(1),
					PasswordHash:    "hash",
				}, nil
			},
		},
		stubPasswordVerifier{
			compareFn: func(string, string) error {
				return errors.New("password mismatch")
			},
		},
		stubTokenManager{},
	)

	_, err := useCase.Login(context.Background(), LoginInput{
		Email:    "alice@example.com",
		Password: "bad-password",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func int64Pointer(value int64) *int64 {
	return &value
}

func TestUseCaseLoginValidatesRequiredFields(t *testing.T) {
	useCase := NewUseCase(stubUserRepository{}, stubPasswordVerifier{}, stubTokenManager{})

	_, err := useCase.Login(context.Background(), LoginInput{})

	var validationErr domainuser.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if validationErr.Field != "email" {
		t.Fatalf("expected email validation error, got %q", validationErr.Field)
	}
}
