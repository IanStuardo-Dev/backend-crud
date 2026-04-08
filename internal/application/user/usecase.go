package userapp

import (
	"context"
	"errors"

	domainuser "github.com/example/crud/internal/domain/user"
)

type useCase struct {
	repo   Repository
	hasher PasswordHasher
}

func NewUseCase(repo Repository, hasher PasswordHasher) UseCase {
	return &useCase{repo: repo, hasher: hasher}
}

func (uc *useCase) Create(ctx context.Context, input CreateInput) (Output, error) {
	user := domainuser.User{
		CompanyID:       cloneInt64Pointer(input.CompanyID),
		Name:            input.Name,
		Email:           input.Email,
		Role:            input.Role,
		IsActive:        input.IsActive,
		DefaultBranchID: cloneInt64Pointer(input.DefaultBranchID),
	}
	user.Normalize()
	if err := user.Validate(); err != nil {
		return Output{}, err
	}
	if err := domainuser.ValidatePassword(input.Password); err != nil {
		return Output{}, err
	}

	passwordHash, err := uc.hasher.Hash(input.Password)
	if err != nil {
		return Output{}, ErrPasswordHashFailed
	}
	user.PasswordHash = passwordHash

	if err := uc.repo.Create(ctx, &user); err != nil {
		return Output{}, err
	}

	return toOutput(user), nil
}

func (uc *useCase) List(ctx context.Context) ([]Output, error) {
	users, err := uc.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	outputs := make([]Output, 0, len(users))
	for _, user := range users {
		outputs = append(outputs, toOutput(user))
	}

	return outputs, nil
}

func (uc *useCase) GetByID(ctx context.Context, id int64) (Output, error) {
	if err := validateID(id); err != nil {
		return Output{}, err
	}

	user, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return Output{}, err
	}
	if user == nil {
		return Output{}, ErrNotFound
	}

	return toOutput(*user), nil
}

func (uc *useCase) Update(ctx context.Context, input UpdateInput) (Output, error) {
	if err := validateID(input.ID); err != nil {
		return Output{}, err
	}

	user := domainuser.User{
		ID:              input.ID,
		CompanyID:       cloneInt64Pointer(input.CompanyID),
		Name:            input.Name,
		Email:           input.Email,
		Role:            input.Role,
		IsActive:        input.IsActive,
		DefaultBranchID: cloneInt64Pointer(input.DefaultBranchID),
	}
	user.Normalize()
	if err := user.Validate(); err != nil {
		return Output{}, err
	}
	if err := uc.repo.Update(ctx, &user); err != nil {
		return Output{}, err
	}

	return toOutput(user), nil
}

func (uc *useCase) Delete(ctx context.Context, id int64) error {
	if err := validateID(id); err != nil {
		return err
	}

	err := uc.repo.Delete(ctx, id)
	if errors.Is(err, ErrNotFound) {
		return ErrNotFound
	}

	return err
}

func validateID(id int64) error {
	if id <= 0 {
		return domainuser.ValidationError{Field: "id", Message: "id must be greater than 0"}
	}

	return nil
}

func toOutput(user domainuser.User) Output {
	return Output{
		ID:              user.ID,
		CompanyID:       cloneInt64Pointer(user.CompanyID),
		Name:            user.Name,
		Email:           user.Email,
		Role:            user.Role,
		IsActive:        user.IsActive,
		DefaultBranchID: cloneInt64Pointer(user.DefaultBranchID),
	}
}

func cloneInt64Pointer(value *int64) *int64 {
	if value == nil {
		return nil
	}

	cloned := *value
	return &cloned
}
