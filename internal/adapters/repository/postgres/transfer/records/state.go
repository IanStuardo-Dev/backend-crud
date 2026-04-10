package transferrecordspg

import (
	"context"

	postgresshared "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/shared"
	domaintransfer "github.com/IanStuardo-Dev/backend-crud/internal/domain/transfer"
)

func (s *Store) Approve(ctx context.Context, transfer *domaintransfer.Transfer) error {
	_, err := postgresshared.Queryer(ctx, s.DB).ExecContext(
		ctx,
		`UPDATE inventory_transfers
		SET status = $1,
			approved_by_user_id = $2,
			approved_at = $3
		WHERE id = $4`,
		transfer.Status,
		transfer.ApprovedByUserID,
		transfer.ApprovedAt,
		transfer.ID,
	)
	return err
}

func (s *Store) Dispatch(ctx context.Context, transfer *domaintransfer.Transfer) error {
	_, err := postgresshared.Queryer(ctx, s.DB).ExecContext(
		ctx,
		`UPDATE inventory_transfers
		SET status = $1,
			dispatched_by_user_id = $2,
			dispatched_at = $3
		WHERE id = $4`,
		transfer.Status,
		transfer.DispatchedByUserID,
		transfer.DispatchedAt,
		transfer.ID,
	)
	return err
}

func (s *Store) Receive(ctx context.Context, transfer *domaintransfer.Transfer) error {
	_, err := postgresshared.Queryer(ctx, s.DB).ExecContext(
		ctx,
		`UPDATE inventory_transfers
		SET status = $1,
			received_by_user_id = $2,
			received_at = $3
		WHERE id = $4`,
		transfer.Status,
		transfer.ReceivedByUserID,
		transfer.ReceivedAt,
		transfer.ID,
	)
	return err
}

func (s *Store) Cancel(ctx context.Context, transfer *domaintransfer.Transfer) error {
	_, err := postgresshared.Queryer(ctx, s.DB).ExecContext(
		ctx,
		`UPDATE inventory_transfers
		SET status = $1,
			cancelled_by_user_id = $2,
			cancelled_at = $3
		WHERE id = $4`,
		transfer.Status,
		transfer.CancelledByUserID,
		transfer.CancelledAt,
		transfer.ID,
	)
	return err
}
