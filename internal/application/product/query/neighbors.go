package productquery

import (
	"context"
	"math"

	productdto "github.com/IanStuardo-Dev/backend-crud/internal/application/product/dto"
	producterrors "github.com/IanStuardo-Dev/backend-crud/internal/application/product/errors"
	productports "github.com/IanStuardo-Dev/backend-crud/internal/application/product/ports"
	domainproduct "github.com/IanStuardo-Dev/backend-crud/internal/domain/product"
)

type FindNeighborsHandler struct {
	reader     productports.ProductCatalogReader
	similarity productports.ProductSimilarityReader
}

func NewFindNeighborsHandler(reader productports.ProductCatalogReader, similarity productports.ProductSimilarityReader) FindNeighborsHandler {
	return FindNeighborsHandler{reader: reader, similarity: similarity}
}

func (h FindNeighborsHandler) Handle(ctx context.Context, input productdto.FindNeighborsInput) (productdto.FindNeighborsOutput, error) {
	if err := validateID(input.ProductID); err != nil {
		return productdto.FindNeighborsOutput{}, err
	}

	limit, err := normalizeLimit(input.Limit)
	if err != nil {
		return productdto.FindNeighborsOutput{}, err
	}
	if err := validateMinSimilarity(input.MinSimilarity); err != nil {
		return productdto.FindNeighborsOutput{}, err
	}

	sourceProduct, err := h.reader.GetByID(ctx, input.ProductID)
	if err != nil {
		return productdto.FindNeighborsOutput{}, err
	}
	if sourceProduct == nil {
		return productdto.FindNeighborsOutput{}, producterrors.ErrNotFound
	}
	if len(sourceProduct.Embedding) == 0 {
		return productdto.FindNeighborsOutput{}, producterrors.ErrSourceEmbeddingMissing
	}

	neighbors, err := h.similarity.FindNeighbors(ctx, sourceProduct.ID, sourceProduct.CompanyID, limit, input.MinSimilarity)
	if err != nil {
		return productdto.FindNeighborsOutput{}, err
	}
	for index := range neighbors {
		neighbors[index].Distance = roundFloat(neighbors[index].Distance, 6)
		neighbors[index].SimilarityPercentage = roundFloat(clamp((1-neighbors[index].Distance)*100, 0, 100), 2)
	}

	return productdto.FindNeighborsOutput{
		SourceProductID:   sourceProduct.ID,
		SourceProductName: sourceProduct.Name,
		SourceCompanyID:   sourceProduct.CompanyID,
		Neighbors:         neighbors,
		Limit:             limit,
	}, nil
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
