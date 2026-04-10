package transferquery

import (
	"context"

	transferdto "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer/dto"
	transfererrors "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer/errors"
	transferports "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer/ports"
	domaintransfer "github.com/IanStuardo-Dev/backend-crud/internal/domain/transfer"
)

type GetByIDHandler struct {
	reader transferports.TransferReader
}

func NewGetByIDHandler(reader transferports.TransferReader) GetByIDHandler {
	return GetByIDHandler{reader: reader}
}

func (h GetByIDHandler) Handle(ctx context.Context, id int64) (transferdto.Output, error) {
	if id <= 0 {
		return transferdto.Output{}, domaintransfer.ValidationError{Field: "id", Message: "id must be greater than 0"}
	}

	transfer, err := h.reader.GetByID(ctx, id)
	if err != nil {
		return transferdto.Output{}, err
	}
	if transfer == nil {
		return transferdto.Output{}, transfererrors.ErrNotFound
	}

	return toOutput(*transfer), nil
}
