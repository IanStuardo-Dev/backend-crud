package productcommand

import (
	"context"

	productdto "github.com/IanStuardo-Dev/backend-crud/internal/application/product/dto"
	producterrors "github.com/IanStuardo-Dev/backend-crud/internal/application/product/errors"
	productports "github.com/IanStuardo-Dev/backend-crud/internal/application/product/ports"
	domainproduct "github.com/IanStuardo-Dev/backend-crud/internal/domain/product"
)

type UpdateHandler struct {
	reader   productports.ProductCatalogReader
	writer   productports.ProductCatalogWriter
	embedder productports.Embedder
}

func NewUpdateHandler(reader productports.ProductCatalogReader, writer productports.ProductCatalogWriter, embedder productports.Embedder) UpdateHandler {
	return UpdateHandler{reader: reader, writer: writer, embedder: embedder}
}

func (h UpdateHandler) Handle(ctx context.Context, input productdto.UpdateInput) (productdto.Output, error) {
	if err := validateID(input.ID); err != nil {
		return productdto.Output{}, err
	}

	existingProduct, err := h.reader.GetByID(ctx, input.ID)
	if err != nil {
		return productdto.Output{}, err
	}
	if existingProduct == nil {
		return productdto.Output{}, producterrors.ErrNotFound
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
	embedding, err := resolveUpdatedEmbedding(ctx, h.embedder, *existingProduct, product, input.Embedding)
	if err != nil {
		return productdto.Output{}, err
	}
	product.Embedding = embedding
	if err := product.Validate(); err != nil {
		return productdto.Output{}, err
	}
	if err := h.writer.Update(ctx, &product); err != nil {
		return productdto.Output{}, err
	}

	return toOutput(product), nil
}
