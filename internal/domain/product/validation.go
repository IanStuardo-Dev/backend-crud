package product

import (
	"fmt"
	"math"
	"strings"
)

// ValidationError represents a business validation failure.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}

	return e.Field + ": " + e.Message
}

func (p *Product) Normalize() {
	p.SKU = strings.TrimSpace(strings.ToUpper(p.SKU))
	p.Name = strings.TrimSpace(p.Name)
	p.Description = strings.TrimSpace(p.Description)
	p.Category = strings.TrimSpace(p.Category)
	p.Brand = strings.TrimSpace(p.Brand)
	p.Currency = strings.TrimSpace(strings.ToUpper(p.Currency))
}

func (p Product) Validate() error {
	if p.CompanyID <= 0 {
		return ValidationError{Field: "company_id", Message: "company_id must be greater than 0"}
	}
	if p.BranchID <= 0 {
		return ValidationError{Field: "branch_id", Message: "branch_id must be greater than 0"}
	}
	if p.SKU == "" {
		return ValidationError{Field: "sku", Message: "sku is required"}
	}
	if len(p.SKU) > 64 {
		return ValidationError{Field: "sku", Message: "sku must be at most 64 characters"}
	}
	if p.Name == "" {
		return ValidationError{Field: "name", Message: "name is required"}
	}
	if p.Category == "" {
		return ValidationError{Field: "category", Message: "category is required"}
	}
	if p.PriceCents < 0 {
		return ValidationError{Field: "price_cents", Message: "price_cents must be greater than or equal to 0"}
	}
	if p.Currency == "" {
		return ValidationError{Field: "currency", Message: "currency is required"}
	}
	if len(p.Currency) != 3 {
		return ValidationError{Field: "currency", Message: "currency must be a 3-letter ISO code"}
	}
	if p.Stock < 0 {
		return ValidationError{Field: "stock", Message: "stock must be greater than or equal to 0"}
	}
	if len(p.Embedding) > 0 {
		if len(p.Embedding) != EmbeddingDimensions {
			return ValidationError{
				Field:   "embedding",
				Message: fmt.Sprintf("embedding must have exactly %d dimensions", EmbeddingDimensions),
			}
		}
		for index, value := range p.Embedding {
			if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
				return ValidationError{
					Field:   "embedding",
					Message: fmt.Sprintf("embedding contains an invalid value at position %d", index),
				}
			}
		}
	}

	return nil
}
