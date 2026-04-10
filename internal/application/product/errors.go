package productapp

import "errors"

var (
	ErrNotFound               = errors.New("product not found")
	ErrConflict               = errors.New("product conflict")
	ErrInvalidReference       = errors.New("invalid product reference")
	ErrSourceEmbeddingMissing = errors.New("source product does not have embedding")
	ErrEmbeddingUnavailable   = errors.New("product embedding provider is not configured")
	ErrEmbeddingGeneration    = errors.New("product embedding generation failed")
	ErrUnauthorizedFeedback   = errors.New("product feedback requires an authenticated user")
)
