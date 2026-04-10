package productquery

import (
	productdto "github.com/IanStuardo-Dev/backend-crud/internal/application/product/dto"
	domainproduct "github.com/IanStuardo-Dev/backend-crud/internal/domain/product"
)

func validateID(id int64) error {
	if id <= 0 {
		return domainproduct.ValidationError{Field: "id", Message: "id must be greater than 0"}
	}

	return nil
}

func cloneEmbedding(embedding []float32) []float32 {
	if len(embedding) == 0 {
		return nil
	}

	cloned := make([]float32, len(embedding))
	copy(cloned, embedding)
	return cloned
}

func toOutput(product domainproduct.Product) productdto.Output {
	return productdto.Output{
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
