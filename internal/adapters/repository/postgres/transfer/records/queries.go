package transferrecordspg

import (
	"context"
	"database/sql"
	"errors"

	domaintransfer "github.com/IanStuardo-Dev/backend-crud/internal/domain/transfer"
)

func (s *Store) List(ctx context.Context) ([]domaintransfer.Transfer, error) {
	return s.listByQuery(ctx, `SELECT id, company_id, origin_branch_id, destination_branch_id, status, requested_by_user_id, supervisor_user_id,
			approved_by_user_id, dispatched_by_user_id, received_by_user_id, cancelled_by_user_id,
			note, created_at, approved_at, dispatched_at, received_at, cancelled_at
		FROM inventory_transfers
		ORDER BY id`)
}

func (s *Store) ListByBranch(ctx context.Context, branchID int64) ([]domaintransfer.Transfer, error) {
	return s.listByQuery(ctx, `SELECT id, company_id, origin_branch_id, destination_branch_id, status, requested_by_user_id, supervisor_user_id,
			approved_by_user_id, dispatched_by_user_id, received_by_user_id, cancelled_by_user_id,
			note, created_at, approved_at, dispatched_at, received_at, cancelled_at
		FROM inventory_transfers
		WHERE origin_branch_id = $1 OR destination_branch_id = $1
		ORDER BY id`, branchID)
}

func (s *Store) GetByID(ctx context.Context, id int64) (*domaintransfer.Transfer, error) {
	row := s.DB.QueryRowContext(
		ctx,
		`SELECT id, company_id, origin_branch_id, destination_branch_id, status, requested_by_user_id, supervisor_user_id,
			approved_by_user_id, dispatched_by_user_id, received_by_user_id, cancelled_by_user_id,
			note, created_at, approved_at, dispatched_at, received_at, cancelled_at
		FROM inventory_transfers
		WHERE id = $1`,
		id,
	)

	transfer, err := scanTransfer(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	transfers, err := attachTransferItems(ctx, s.DB, []domaintransfer.Transfer{transfer})
	if err != nil {
		return nil, err
	}
	return &transfers[0], nil
}

func (s *Store) listByQuery(ctx context.Context, query string, args ...any) ([]domaintransfer.Transfer, error) {
	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transfers := make([]domaintransfer.Transfer, 0)
	for rows.Next() {
		transfer, err := scanTransfer(rows)
		if err != nil {
			return nil, err
		}
		transfers = append(transfers, transfer)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return attachTransferItems(ctx, s.DB, transfers)
}
