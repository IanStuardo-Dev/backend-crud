package userapp

import "errors"

var (
	ErrNotFound           = errors.New("user not found")
	ErrConflict           = errors.New("user conflict")
	ErrPasswordHashFailed = errors.New("password hashing failed")
)
