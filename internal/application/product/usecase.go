package productapp

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	domainproduct "github.com/IanStuardo-Dev/backend-crud/internal/domain/product"
)

type useCase struct {
	repo     Repository
	embedder Embedder
}

func NewUseCase(repo Repository, embedder Embedder) UseCase {
	return &useCase{
		repo:     repo,
		embedder: embedder,
	}
}

func (uc *useCase) Create(ctx context.Context, input CreateInput) (Output, error) {
	embedding, err := uc.resolveEmbedding(ctx, productTextSource{
		SKU:         input.SKU,
		Name:        input.Name,
		Description: input.Description,
		Category:    input.Category,
		Brand:       input.Brand,
		Embedding:   input.Embedding,
	})
	if err != nil {
		return Output{}, err
	}

	product := domainproduct.Product{
		CompanyID:   input.CompanyID,
		BranchID:    input.BranchID,
		SKU:         input.SKU,
		Name:        input.Name,
		Description: input.Description,
		Category:    input.Category,
		Brand:       input.Brand,
		PriceCents:  input.PriceCents,
		Currency:    input.Currency,
		Stock:       input.Stock,
		Embedding:   embedding,
	}
	product.Normalize()
	if err := product.Validate(); err != nil {
		return Output{}, err
	}
	if err := uc.repo.Create(ctx, &product); err != nil {
		return Output{}, err
	}

	return toOutput(product), nil
}

func (uc *useCase) List(ctx context.Context) ([]Output, error) {
	products, err := uc.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	outputs := make([]Output, 0, len(products))
	for _, product := range products {
		outputs = append(outputs, toOutput(product))
	}

	return outputs, nil
}

func (uc *useCase) GetByID(ctx context.Context, id int64) (Output, error) {
	if err := validateID(id); err != nil {
		return Output{}, err
	}

	product, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return Output{}, err
	}
	if product == nil {
		return Output{}, ErrNotFound
	}

	return toOutput(*product), nil
}

func (uc *useCase) FindNeighbors(ctx context.Context, input FindNeighborsInput) (FindNeighborsOutput, error) {
	if err := validateID(input.ProductID); err != nil {
		return FindNeighborsOutput{}, err
	}

	limit, err := normalizeLimit(input.Limit)
	if err != nil {
		return FindNeighborsOutput{}, err
	}
	if err := validateMinSimilarity(input.MinSimilarity); err != nil {
		return FindNeighborsOutput{}, err
	}

	sourceProduct, err := uc.repo.GetByID(ctx, input.ProductID)
	if err != nil {
		return FindNeighborsOutput{}, err
	}
	if sourceProduct == nil {
		return FindNeighborsOutput{}, ErrNotFound
	}
	if len(sourceProduct.Embedding) == 0 {
		return FindNeighborsOutput{}, ErrSourceEmbeddingMissing
	}

	neighbors, err := uc.repo.FindNeighbors(ctx, sourceProduct.ID, sourceProduct.CompanyID, limit, input.MinSimilarity)
	if err != nil {
		return FindNeighborsOutput{}, err
	}

	for index := range neighbors {
		neighbors[index].Distance = roundFloat(neighbors[index].Distance, 6)
		neighbors[index].SimilarityPercentage = roundFloat(clamp((1-neighbors[index].Distance)*100, 0, 100), 2)
	}

	return FindNeighborsOutput{
		SourceProductID:   sourceProduct.ID,
		SourceProductName: sourceProduct.Name,
		SourceCompanyID:   sourceProduct.CompanyID,
		Neighbors:         neighbors,
		Limit:             limit,
	}, nil
}

func (uc *useCase) RecordNeighborFeedback(ctx context.Context, input RecordNeighborFeedbackInput) (NeighborFeedbackOutput, error) {
	if err := validateID(input.SourceProductID); err != nil {
		return NeighborFeedbackOutput{}, err
	}
	if err := validateNeighborFeedbackID("suggested_product_id", input.SuggestedProductID); err != nil {
		return NeighborFeedbackOutput{}, err
	}
	if err := validateNeighborFeedbackID("branch_id", input.BranchID); err != nil {
		return NeighborFeedbackOutput{}, err
	}
	if err := validateNeighborFeedbackID("user_id", input.UserID); err != nil {
		return NeighborFeedbackOutput{}, err
	}
	if input.SourceProductID == input.SuggestedProductID {
		return NeighborFeedbackOutput{}, domainproduct.ValidationError{
			Field:   "suggested_product_id",
			Message: "suggested_product_id must be different from the source product",
		}
	}

	action, err := normalizeNeighborFeedbackAction(input.Action)
	if err != nil {
		return NeighborFeedbackOutput{}, err
	}

	note, err := normalizeNeighborFeedbackNote(input.Note)
	if err != nil {
		return NeighborFeedbackOutput{}, err
	}

	sourceProduct, err := uc.repo.GetByID(ctx, input.SourceProductID)
	if err != nil {
		return NeighborFeedbackOutput{}, err
	}
	if sourceProduct == nil {
		return NeighborFeedbackOutput{}, ErrNotFound
	}

	suggestedProduct, err := uc.repo.GetByID(ctx, input.SuggestedProductID)
	if err != nil {
		return NeighborFeedbackOutput{}, err
	}
	if suggestedProduct == nil {
		return NeighborFeedbackOutput{}, ErrNotFound
	}
	if sourceProduct.CompanyID != suggestedProduct.CompanyID {
		return NeighborFeedbackOutput{}, ErrInvalidReference
	}

	return uc.repo.SaveNeighborFeedback(ctx, RecordNeighborFeedbackInput{
		SourceProductID:    input.SourceProductID,
		SuggestedProductID: input.SuggestedProductID,
		CompanyID:          sourceProduct.CompanyID,
		BranchID:           input.BranchID,
		UserID:             input.UserID,
		Action:             action,
		Note:               note,
	})
}

func (uc *useCase) Update(ctx context.Context, input UpdateInput) (Output, error) {
	if err := validateID(input.ID); err != nil {
		return Output{}, err
	}

	existingProduct, err := uc.repo.GetByID(ctx, input.ID)
	if err != nil {
		return Output{}, err
	}
	if existingProduct == nil {
		return Output{}, ErrNotFound
	}

	product := domainproduct.Product{
		ID:          input.ID,
		CompanyID:   input.CompanyID,
		BranchID:    input.BranchID,
		SKU:         input.SKU,
		Name:        input.Name,
		Description: input.Description,
		Category:    input.Category,
		Brand:       input.Brand,
		PriceCents:  input.PriceCents,
		Currency:    input.Currency,
		Stock:       input.Stock,
	}
	product.Normalize()
	embedding, err := uc.resolveUpdatedEmbedding(ctx, *existingProduct, product, input.Embedding)
	if err != nil {
		return Output{}, err
	}
	product.Embedding = embedding
	if err := product.Validate(); err != nil {
		return Output{}, err
	}
	if err := uc.repo.Update(ctx, &product); err != nil {
		return Output{}, err
	}

	return toOutput(product), nil
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
		return domainproduct.ValidationError{Field: "id", Message: "id must be greater than 0"}
	}

	return nil
}

func toOutput(product domainproduct.Product) Output {
	return Output{
		ID:          product.ID,
		CompanyID:   product.CompanyID,
		BranchID:    product.BranchID,
		SKU:         product.SKU,
		Name:        product.Name,
		Description: product.Description,
		Category:    product.Category,
		Brand:       product.Brand,
		PriceCents:  product.PriceCents,
		Currency:    product.Currency,
		Stock:       product.Stock,
		Embedding:   cloneEmbedding(product.Embedding),
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
}

func cloneEmbedding(embedding []float32) []float32 {
	if len(embedding) == 0 {
		return nil
	}

	cloned := make([]float32, len(embedding))
	copy(cloned, embedding)
	return cloned
}

func normalizeLimit(limit int) (int, error) {
	if limit == 0 {
		return 5, nil
	}
	if limit < 0 {
		return 0, domainproduct.ValidationError{Field: "limit", Message: "limit must be greater than 0"}
	}
	if limit > 10 {
		return 0, domainproduct.ValidationError{Field: "limit", Message: "limit must be less than or equal to 10"}
	}

	return limit, nil
}

func validateMinSimilarity(minSimilarity float64) error {
	if minSimilarity < 0 || minSimilarity > 1 {
		return domainproduct.ValidationError{Field: "min_similarity", Message: "min_similarity must be between 0 and 1"}
	}

	return nil
}

func validateNeighborFeedbackID(field string, id int64) error {
	if id <= 0 {
		return domainproduct.ValidationError{Field: field, Message: field + " must be greater than 0"}
	}

	return nil
}

func normalizeNeighborFeedbackAction(value string) (string, error) {
	action := strings.ToLower(strings.TrimSpace(value))
	switch action {
	case "accepted", "rejected", "ignored":
		return action, nil
	default:
		return "", domainproduct.ValidationError{
			Field:   "action",
			Message: "action must be one of accepted, rejected, or ignored",
		}
	}
}

func normalizeNeighborFeedbackNote(value string) (string, error) {
	note := strings.TrimSpace(value)
	if len(note) > 1000 {
		return "", domainproduct.ValidationError{
			Field:   "note",
			Message: "note must be less than or equal to 1000 characters",
		}
	}

	return note, nil
}

type productTextSource struct {
	SKU         string
	Name        string
	Description string
	Category    string
	Brand       string
	Embedding   []float32
}

func (uc *useCase) resolveEmbedding(ctx context.Context, source productTextSource) ([]float32, error) {
	if len(source.Embedding) > 0 {
		return cloneEmbedding(source.Embedding), nil
	}
	if uc.embedder == nil {
		return nil, ErrEmbeddingUnavailable
	}

	embedding, err := uc.embedder.EmbedText(ctx, buildEmbeddingText(source))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEmbeddingGeneration, err)
	}

	return cloneEmbedding(embedding), nil
}

func (uc *useCase) resolveUpdatedEmbedding(ctx context.Context, existing, candidate domainproduct.Product, provided []float32) ([]float32, error) {
	if len(provided) > 0 {
		return cloneEmbedding(provided), nil
	}

	if buildEmbeddingText(productTextSourceFromProduct(existing)) == buildEmbeddingText(productTextSourceFromProduct(candidate)) {
		return cloneEmbedding(existing.Embedding), nil
	}

	return uc.resolveEmbedding(ctx, productTextSourceFromProduct(candidate))
}

func buildEmbeddingText(source productTextSource) string {
	parts := make([]string, 0, 5)
	if value := strings.TrimSpace(source.SKU); value != "" {
		parts = append(parts, "SKU: "+value)
	}
	if value := strings.TrimSpace(source.Name); value != "" {
		parts = append(parts, "Name: "+value)
	}
	if value := strings.TrimSpace(source.Description); value != "" {
		parts = append(parts, "Description: "+value)
	}
	if value := strings.TrimSpace(source.Category); value != "" {
		parts = append(parts, "Category: "+value)
	}
	if value := strings.TrimSpace(source.Brand); value != "" {
		parts = append(parts, "Brand: "+value)
	}

	return strings.Join(parts, "\n")
}

func productTextSourceFromProduct(product domainproduct.Product) productTextSource {
	return productTextSource{
		SKU:         product.SKU,
		Name:        product.Name,
		Description: product.Description,
		Category:    product.Category,
		Brand:       product.Brand,
		Embedding:   product.Embedding,
	}
}

func roundFloat(value float64, decimals int) float64 {
	factor := math.Pow(10, float64(decimals))
	return math.Round(value*factor) / factor
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
