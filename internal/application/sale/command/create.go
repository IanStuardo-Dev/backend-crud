package salecommand

import (
	"context"

	saledto "github.com/IanStuardo-Dev/backend-crud/internal/application/sale/dto"
	saleerrors "github.com/IanStuardo-Dev/backend-crud/internal/application/sale/errors"
	saleports "github.com/IanStuardo-Dev/backend-crud/internal/application/sale/ports"
	domainsale "github.com/IanStuardo-Dev/backend-crud/internal/domain/sale"
)

type CreateHandler struct {
	txManager      saleports.TransactionManager
	saleWriter     saleports.SaleWriter
	stockReader    saleports.InventoryStockReader
	stockWriter    saleports.InventoryStockWriter
	movementWriter saleports.InventoryMovementWriter
	branchReader   saleports.BranchReferenceReader
	userReader     saleports.UserReferenceReader
}

func NewCreateHandler(
	txManager saleports.TransactionManager,
	saleWriter saleports.SaleWriter,
	stockReader saleports.InventoryStockReader,
	stockWriter saleports.InventoryStockWriter,
	movementWriter saleports.InventoryMovementWriter,
	branchReader saleports.BranchReferenceReader,
	userReader saleports.UserReferenceReader,
) CreateHandler {
	return CreateHandler{
		txManager:      txManager,
		saleWriter:     saleWriter,
		stockReader:    stockReader,
		stockWriter:    stockWriter,
		movementWriter: movementWriter,
		branchReader:   branchReader,
		userReader:     userReader,
	}
}

func (h CreateHandler) Handle(ctx context.Context, input saledto.CreateInput) (saledto.Output, error) {
	sale := domainsale.Sale{
		CompanyID:       input.CompanyID,
		BranchID:        input.BranchID,
		CreatedByUserID: input.CreatedByUserID,
		Items:           toDomainItems(input.Items),
	}
	if err := sale.ValidateForCreate(); err != nil {
		return saledto.Output{}, err
	}

	err := h.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		branchExists, err := h.branchReader.BranchExists(txCtx, sale.CompanyID, sale.BranchID)
		if err != nil {
			return err
		}
		if !branchExists {
			return saleerrors.ErrInvalidReference
		}

		userExists, err := h.userReader.UserExists(txCtx, sale.CreatedByUserID)
		if err != nil {
			return err
		}
		if !userExists {
			return saleerrors.ErrInvalidReference
		}

		productRows, err := h.stockReader.LoadForSale(txCtx, sale.CompanyID, sale.BranchID, sale.Items, true)
		if err != nil {
			return err
		}

		var totalAmount int64
		for index, item := range sale.Items {
			product, ok := productRows[item.ProductID]
			if !ok {
				return saleerrors.ErrInvalidReference
			}
			if product.AvailableStock < item.Quantity {
				return saleerrors.ErrInsufficientStock
			}

			sale.Items[index].ProductSKU = product.SKU
			sale.Items[index].ProductName = product.Name
			sale.Items[index].UnitPriceCents = product.PriceCents
			sale.Items[index].SubtotalCents = int64(item.Quantity) * product.PriceCents
			totalAmount += sale.Items[index].SubtotalCents
		}

		sale.TotalAmountCents = totalAmount
		if err := h.saleWriter.Create(txCtx, &sale); err != nil {
			return err
		}

		for _, item := range sale.Items {
			product := productRows[item.ProductID]
			stockAfter := product.StockOnHand - item.Quantity
			if err := h.stockWriter.SetStockOnHand(txCtx, sale.CompanyID, sale.BranchID, item.ProductID, stockAfter); err != nil {
				return err
			}
			if err := h.movementWriter.CreateSaleMovement(txCtx, saleports.MovementInput{
				CompanyID:       sale.CompanyID,
				BranchID:        sale.BranchID,
				ProductID:       item.ProductID,
				SaleID:          sale.ID,
				QuantityDelta:   -item.Quantity,
				StockAfter:      stockAfter,
				CreatedByUserID: sale.CreatedByUserID,
			}); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return saledto.Output{}, err
	}

	return toOutput(sale), nil
}
