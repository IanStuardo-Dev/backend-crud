package saleapp

import "errors"

var (
	ErrNotFound          = errors.New("sale not found")
	ErrInvalidReference  = errors.New("invalid sale reference")
	ErrInsufficientStock = errors.New("insufficient stock")
)
