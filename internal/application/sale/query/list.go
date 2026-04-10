package salequery

import (
	"context"

	saledto "github.com/IanStuardo-Dev/backend-crud/internal/application/sale/dto"
	saleports "github.com/IanStuardo-Dev/backend-crud/internal/application/sale/ports"
)

type ListHandler struct {
	reader saleports.SaleReader
}

func NewListHandler(reader saleports.SaleReader) ListHandler {
	return ListHandler{reader: reader}
}

func (h ListHandler) Handle(ctx context.Context) ([]saledto.Output, error) {
	sales, err := h.reader.List(ctx)
	if err != nil {
		return nil, err
	}

	outputs := make([]saledto.Output, 0, len(sales))
	for _, sale := range sales {
		outputs = append(outputs, toOutput(sale))
	}

	return outputs, nil
}
