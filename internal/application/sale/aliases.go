package saleapp

import (
	saledto "github.com/IanStuardo-Dev/backend-crud/internal/application/sale/dto"
	saleerrors "github.com/IanStuardo-Dev/backend-crud/internal/application/sale/errors"
	saleports "github.com/IanStuardo-Dev/backend-crud/internal/application/sale/ports"
)

type CreateInput = saledto.CreateInput
type CreateItemInput = saledto.CreateItemInput
type Output = saledto.Output
type ItemOutput = saledto.ItemOutput

type SaleWriter = saleports.SaleWriter
type SaleReader = saleports.SaleReader
type BranchReferenceReader = saleports.BranchReferenceReader
type UserReferenceReader = saleports.UserReferenceReader
type StockSnapshot = saleports.StockSnapshot
type InventoryStockReader = saleports.InventoryStockReader
type InventoryStockWriter = saleports.InventoryStockWriter
type MovementInput = saleports.MovementInput
type InventoryMovementWriter = saleports.InventoryMovementWriter
type TransactionManager = saleports.TransactionManager

var (
	ErrNotFound          = saleerrors.ErrNotFound
	ErrInvalidReference  = saleerrors.ErrInvalidReference
	ErrInsufficientStock = saleerrors.ErrInsufficientStock
)
