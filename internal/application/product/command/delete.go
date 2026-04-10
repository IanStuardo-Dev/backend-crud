package productcommand

import (
	"context"
	"errors"

	producterrors "github.com/IanStuardo-Dev/backend-crud/internal/application/product/errors"
	productports "github.com/IanStuardo-Dev/backend-crud/internal/application/product/ports"
)

type DeleteHandler struct {
	writer productports.ProductCatalogWriter
}

func NewDeleteHandler(writer productports.ProductCatalogWriter) DeleteHandler {
	return DeleteHandler{writer: writer}
}

func (h DeleteHandler) Handle(ctx context.Context, id int64) error {
	if err := validateID(id); err != nil {
		return err
	}

	err := h.writer.Delete(ctx, id)
	if errors.Is(err, producterrors.ErrNotFound) {
		return producterrors.ErrNotFound
	}

	return err
}
