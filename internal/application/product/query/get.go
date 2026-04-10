package productquery

import (
	"context"

	productdto "github.com/IanStuardo-Dev/backend-crud/internal/application/product/dto"
	producterrors "github.com/IanStuardo-Dev/backend-crud/internal/application/product/errors"
	productports "github.com/IanStuardo-Dev/backend-crud/internal/application/product/ports"
)

type GetByIDHandler struct {
	reader productports.ProductCatalogReader
}

func NewGetByIDHandler(reader productports.ProductCatalogReader) GetByIDHandler {
	return GetByIDHandler{reader: reader}
}

func (h GetByIDHandler) Handle(ctx context.Context, id int64) (productdto.Output, error) {
	if err := validateID(id); err != nil {
		return productdto.Output{}, err
	}

	product, err := h.reader.GetByID(ctx, id)
	if err != nil {
		return productdto.Output{}, err
	}
	if product == nil {
		return productdto.Output{}, producterrors.ErrNotFound
	}

	return toOutput(*product), nil
}
