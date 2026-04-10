package transferrecordspg

import (
	"context"

	postgresshared "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/shared"
	transferapp "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer"
	domaintransfer "github.com/IanStuardo-Dev/backend-crud/internal/domain/transfer"
)

func (s *Store) Create(ctx context.Context, transfer *domaintransfer.Transfer) error {
	err := postgresshared.Queryer(ctx, s.DB).QueryRowContext(
		ctx,
		`INSERT INTO inventory_transfers (
			company_id,
			origin_branch_id,
			destination_branch_id,
			status,
			requested_by_user_id,
			supervisor_user_id,
			note,
			created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,NOW())
		RETURNING id, created_at`,
		transfer.CompanyID,
		transfer.OriginBranchID,
		transfer.DestinationBranchID,
		transfer.Status,
		transfer.RequestedByUserID,
		transfer.SupervisorUserID,
		transfer.Note,
	).Scan(&transfer.ID, &transfer.CreatedAt)
	if postgresshared.IsForeignKeyViolation(err) {
		return transferapp.ErrInvalidReference
	}
	if err != nil {
		return err
	}

	for _, item := range transfer.Items {
		if _, err := postgresshared.Queryer(ctx, s.DB).ExecContext(
			ctx,
			`INSERT INTO inventory_transfer_items (transfer_id, product_id, quantity)
			VALUES ($1,$2,$3)`,
			transfer.ID,
			item.ProductID,
			item.Quantity,
		); err != nil {
			return err
		}
	}

	return nil
}
