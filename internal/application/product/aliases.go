package productapp

import (
	productdto "github.com/IanStuardo-Dev/backend-crud/internal/application/product/dto"
	producterrors "github.com/IanStuardo-Dev/backend-crud/internal/application/product/errors"
	productports "github.com/IanStuardo-Dev/backend-crud/internal/application/product/ports"
)

type CreateInput = productdto.CreateInput
type UpdateInput = productdto.UpdateInput
type Output = productdto.Output
type FindNeighborsInput = productdto.FindNeighborsInput
type NeighborOutput = productdto.NeighborOutput
type FindNeighborsOutput = productdto.FindNeighborsOutput
type RecordNeighborFeedbackInput = productdto.RecordNeighborFeedbackInput
type NeighborFeedbackOutput = productdto.NeighborFeedbackOutput

type ProductCatalogWriter = productports.ProductCatalogWriter
type ProductCatalogReader = productports.ProductCatalogReader
type ProductSimilarityReader = productports.ProductSimilarityReader
type ProductSuggestionFeedbackWriter = productports.ProductSuggestionFeedbackWriter
type Embedder = productports.Embedder

var (
	ErrNotFound               = producterrors.ErrNotFound
	ErrConflict               = producterrors.ErrConflict
	ErrInvalidReference       = producterrors.ErrInvalidReference
	ErrSourceEmbeddingMissing = producterrors.ErrSourceEmbeddingMissing
	ErrEmbeddingUnavailable   = producterrors.ErrEmbeddingUnavailable
	ErrEmbeddingGeneration    = producterrors.ErrEmbeddingGeneration
	ErrUnauthorizedFeedback   = producterrors.ErrUnauthorizedFeedback
)
