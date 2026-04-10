package postgressale

import (
	"database/sql"

	salemovementspg "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/sale/movements"
	salerecordspg "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/sale/records"
	salereferencespg "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/sale/references"
	salestockpg "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/sale/stock"
)

func NewSaleStore(db *sql.DB) *salerecordspg.Store {
	return salerecordspg.NewStore(db)
}

func NewStockStore(db *sql.DB) *salestockpg.Store {
	return salestockpg.NewStore(db)
}

func NewMovementStore(db *sql.DB) *salemovementspg.Store {
	return salemovementspg.NewStore(db)
}

func NewReferenceStore(db *sql.DB) *salereferencespg.Store {
	return salereferencespg.NewStore(db)
}
