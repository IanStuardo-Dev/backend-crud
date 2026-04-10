package transfermovementspg

import (
	"context"
	"database/sql"

	postgresshared "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/shared"
	transferports "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer/ports"
)

type Store struct {
	DB *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{DB: db}
}

func (s *Store) CreateTransferMovement(ctx context.Context, input transferports.MovementInput) error {
	_, err := postgresshared.Queryer(ctx, s.DB).ExecContext(
		ctx,
		`INSERT INTO inventory_movements (
			company_id,
			branch_id,
			product_id,
			sale_id,
			transfer_id,
			movement_type,
			quantity_delta,
			stock_after,
			created_by_user_id,
			created_at
		) VALUES ($1,$2,$3,NULL,$4,$5,$6,$7,$8,NOW())`,
		input.CompanyID,
		input.BranchID,
		input.ProductID,
		input.TransferID,
		input.MovementType,
		input.QuantityDelta,
		input.StockAfter,
		input.CreatedByUserID,
	)
	return err
}
