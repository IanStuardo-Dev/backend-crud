package transferquery

import (
	"context"

	transferdto "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer/dto"
	transferports "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer/ports"
	domaintransfer "github.com/IanStuardo-Dev/backend-crud/internal/domain/transfer"
)

type ListHandler struct {
	reader transferports.TransferReader
}

func NewListHandler(reader transferports.TransferReader) ListHandler {
	return ListHandler{reader: reader}
}

func (h ListHandler) Handle(ctx context.Context) ([]transferdto.Output, error) {
	transfers, err := h.reader.List(ctx)
	if err != nil {
		return nil, err
	}
	return toOutputs(transfers), nil
}

type ListByBranchHandler struct {
	reader transferports.TransferReader
}

func NewListByBranchHandler(reader transferports.TransferReader) ListByBranchHandler {
	return ListByBranchHandler{reader: reader}
}

func (h ListByBranchHandler) Handle(ctx context.Context, branchID int64) ([]transferdto.Output, error) {
	if branchID <= 0 {
		return nil, domaintransfer.ValidationError{Field: "branch_id", Message: "branch_id must be greater than 0"}
	}

	transfers, err := h.reader.ListByBranch(ctx, branchID)
	if err != nil {
		return nil, err
	}
	return toOutputs(transfers), nil
}
