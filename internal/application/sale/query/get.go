package salequery

import (
	"context"

	saledto "github.com/IanStuardo-Dev/backend-crud/internal/application/sale/dto"
	saleerrors "github.com/IanStuardo-Dev/backend-crud/internal/application/sale/errors"
	saleports "github.com/IanStuardo-Dev/backend-crud/internal/application/sale/ports"
	domainsale "github.com/IanStuardo-Dev/backend-crud/internal/domain/sale"
)

type GetByIDHandler struct {
	reader saleports.SaleReader
}

func NewGetByIDHandler(reader saleports.SaleReader) GetByIDHandler {
	return GetByIDHandler{reader: reader}
}

func (h GetByIDHandler) Handle(ctx context.Context, id int64) (saledto.Output, error) {
	if id <= 0 {
		return saledto.Output{}, domainsale.ValidationError{Field: "id", Message: "id must be greater than 0"}
	}

	sale, err := h.reader.GetByID(ctx, id)
	if err != nil {
		return saledto.Output{}, err
	}
	if sale == nil {
		return saledto.Output{}, saleerrors.ErrNotFound
	}

	return toOutput(*sale), nil
}
