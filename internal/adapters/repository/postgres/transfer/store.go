package postgrestransfer

import (
	"database/sql"

	transfermovementspg "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/transfer/movements"
	transferrecordspg "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/transfer/records"
	transferreferencespg "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/transfer/references"
	transferstockpg "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/transfer/stock"
)

func NewTransferStore(db *sql.DB) *transferrecordspg.Store {
	return transferrecordspg.NewStore(db)
}

func NewStockStore(db *sql.DB) *transferstockpg.Store {
	return transferstockpg.NewStore(db)
}

func NewMovementStore(db *sql.DB) *transfermovementspg.Store {
	return transfermovementspg.NewStore(db)
}

func NewReferenceStore(db *sql.DB) *transferreferencespg.Store {
	return transferreferencespg.NewStore(db)
}
