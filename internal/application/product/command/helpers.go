package productcommand

import (
	"context"
	"fmt"
	"strings"

	productdto "github.com/IanStuardo-Dev/backend-crud/internal/application/product/dto"
	producterrors "github.com/IanStuardo-Dev/backend-crud/internal/application/product/errors"
	productports "github.com/IanStuardo-Dev/backend-crud/internal/application/product/ports"
	domainproduct "github.com/IanStuardo-Dev/backend-crud/internal/domain/product"
)

type productTextSource struct {
	SKU         string
	Name        string
	Description string
	Category    string
	Brand       string
	Embedding   []float32
}

func cloneEmbedding(embedding []float32) []float32 {
	if len(embedding) == 0 {
		return nil
	}

	cloned := make([]float32, len(embedding))
	copy(cloned, embedding)
	return cloned
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

func createTextSource(input productdto.CreateInput) productTextSource {
	return productTextSource{
		SKU:         input.SKU,
		Name:        input.Name,
		Description: input.Description,
		Category:    input.Category,
		Brand:       input.Brand,
		Embedding:   input.Embedding,
	}
}

func resolveEmbedding(ctx context.Context, embedder productports.Embedder, source productTextSource) ([]float32, error) {
	if len(source.Embedding) > 0 {
		return cloneEmbedding(source.Embedding), nil
	}
	if embedder == nil {
		return nil, producterrors.ErrEmbeddingUnavailable
	}

	embedding, err := embedder.EmbedText(ctx, buildEmbeddingText(source))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", producterrors.ErrEmbeddingGeneration, err)
	}

	return cloneEmbedding(embedding), nil
}

func resolveUpdatedEmbedding(ctx context.Context, embedder productports.Embedder, existing, candidate domainproduct.Product, provided []float32) ([]float32, error) {
	if len(provided) > 0 {
		return cloneEmbedding(provided), nil
	}
	if buildEmbeddingText(productTextSourceFromProduct(existing)) == buildEmbeddingText(productTextSourceFromProduct(candidate)) {
		return cloneEmbedding(existing.Embedding), nil
	}

	return resolveEmbedding(ctx, embedder, productTextSourceFromProduct(candidate))
}
