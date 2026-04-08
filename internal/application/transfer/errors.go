package transferapp

import "errors"

var (
	ErrNotFound          = errors.New("transfer not found")
	ErrInvalidReference  = errors.New("invalid transfer reference")
	ErrInsufficientStock = errors.New("insufficient stock")
)
