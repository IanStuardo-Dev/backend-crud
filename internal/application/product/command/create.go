package productcommand

import (
	"context"

	productdto "github.com/IanStuardo-Dev/backend-crud/internal/application/product/dto"
	productports "github.com/IanStuardo-Dev/backend-crud/internal/application/product/ports"
	domainproduct "github.com/IanStuardo-Dev/backend-crud/internal/domain/product"
)

type CreateHandler struct {
	writer   productports.ProductCatalogWriter
	embedder productports.Embedder
}

func NewCreateHandler(writer productports.ProductCatalogWriter, embedder productports.Embedder) CreateHandler {
	return CreateHandler{writer: writer, embedder: embedder}
}

func (h CreateHandler) Handle(ctx context.Context, input productdto.CreateInput) (productdto.Output, error) {
	embedding, err := resolveEmbedding(ctx, h.embedder, createTextSource(input))
	if err != nil {
		return productdto.Output{}, err
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
		return productdto.Output{}, err
	}
	if err := h.writer.Create(ctx, &product); err != nil {
		return productdto.Output{}, err
	}

	return toOutput(product), nil
}
