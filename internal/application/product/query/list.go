package productquery

import (
	"context"

	productdto "github.com/IanStuardo-Dev/backend-crud/internal/application/product/dto"
	productports "github.com/IanStuardo-Dev/backend-crud/internal/application/product/ports"
)

type ListHandler struct {
	reader productports.ProductCatalogReader
}

func NewListHandler(reader productports.ProductCatalogReader) ListHandler {
	return ListHandler{reader: reader}
}

func (h ListHandler) Handle(ctx context.Context) ([]productdto.Output, error) {
	products, err := h.reader.List(ctx)
	if err != nil {
		return nil, err
	}

	outputs := make([]productdto.Output, 0, len(products))
	for _, product := range products {
		outputs = append(outputs, toOutput(product))
	}

	return outputs, nil
}
