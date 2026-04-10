package transfererrors

import "errors"

var (
	ErrNotFound          = errors.New("transfer not found")
	ErrInvalidReference  = errors.New("invalid transfer reference")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrForbiddenAction   = errors.New("forbidden transfer action")
	ErrInvalidState      = errors.New("invalid transfer state")
)
