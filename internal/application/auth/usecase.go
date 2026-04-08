package authapp

import (
	"context"
	"strings"

	domainuser "github.com/IanStuardo-Dev/backend-crud/internal/domain/user"
)

type useCase struct {
	users    UserRepository
	password PasswordVerifier
	tokens   TokenManager
}

func NewUseCase(users UserRepository, password PasswordVerifier, tokens TokenManager) UseCase {
	return &useCase{
		users:    users,
		password: password,
		tokens:   tokens,
	}
}

func (uc *useCase) Login(ctx context.Context, input LoginInput) (LoginOutput, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if email == "" {
		return LoginOutput{}, domainuser.ValidationError{Field: "email", Message: "email is required"}
	}
	if strings.TrimSpace(input.Password) == "" {
		return LoginOutput{}, domainuser.ValidationError{Field: "password", Message: "password is required"}
	}

	user, err := uc.users.GetByEmail(ctx, email)
	if err != nil {
		return LoginOutput{}, err
	}
	if user == nil || user.PasswordHash == "" {
		return LoginOutput{}, ErrInvalidCredentials
	}
	if !user.IsActive {
		return LoginOutput{}, ErrInactiveUser
	}
	if err := uc.password.Compare(user.PasswordHash, input.Password); err != nil {
		return LoginOutput{}, ErrInvalidCredentials
	}

	authenticatedUser := AuthenticatedUser{
		ID:              user.ID,
		CompanyID:       cloneInt64Pointer(user.CompanyID),
		Name:            user.Name,
		Email:           user.Email,
		Role:            user.Role,
		IsActive:        user.IsActive,
		DefaultBranchID: cloneInt64Pointer(user.DefaultBranchID),
	}
	token, err := uc.tokens.Generate(authenticatedUser)
	if err != nil {
		return LoginOutput{}, err
	}

	return LoginOutput{
		AccessToken: token.AccessToken,
		TokenType:   "Bearer",
		ExpiresAt:   token.ExpiresAt,
		User: UserOutput{
			ID:              user.ID,
			CompanyID:       cloneInt64Pointer(user.CompanyID),
			Name:            user.Name,
			Email:           user.Email,
			Role:            user.Role,
			IsActive:        user.IsActive,
			DefaultBranchID: cloneInt64Pointer(user.DefaultBranchID),
		},
	}, nil
}

func cloneInt64Pointer(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
