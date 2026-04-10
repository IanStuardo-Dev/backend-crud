package transferrecordspg

import (
	"context"
	"database/sql"

	domaintransfer "github.com/IanStuardo-Dev/backend-crud/internal/domain/transfer"
	"github.com/lib/pq"
)

type transferScanner interface {
	Scan(dest ...any) error
}

func scanTransfer(row transferScanner) (domaintransfer.Transfer, error) {
	var (
		transfer               domaintransfer.Transfer
		approvedByUserIDNull   sql.NullInt64
		dispatchedByUserIDNull sql.NullInt64
		receivedByUserIDNull   sql.NullInt64
		cancelledByUserIDNull  sql.NullInt64
		approvedAtNull         sql.NullTime
		dispatchedAtNull       sql.NullTime
		receivedAtNull         sql.NullTime
		cancelledAtNull        sql.NullTime
	)
	err := row.Scan(
		&transfer.ID,
		&transfer.CompanyID,
		&transfer.OriginBranchID,
		&transfer.DestinationBranchID,
		&transfer.Status,
		&transfer.RequestedByUserID,
		&transfer.SupervisorUserID,
		&approvedByUserIDNull,
		&dispatchedByUserIDNull,
		&receivedByUserIDNull,
		&cancelledByUserIDNull,
		&transfer.Note,
		&transfer.CreatedAt,
		&approvedAtNull,
		&dispatchedAtNull,
		&receivedAtNull,
		&cancelledAtNull,
	)
	if err != nil {
		return domaintransfer.Transfer{}, err
	}
	if approvedByUserIDNull.Valid {
		transfer.ApprovedByUserID = &approvedByUserIDNull.Int64
	}
	if dispatchedByUserIDNull.Valid {
		transfer.DispatchedByUserID = &dispatchedByUserIDNull.Int64
	}
	if receivedByUserIDNull.Valid {
		transfer.ReceivedByUserID = &receivedByUserIDNull.Int64
	}
	if cancelledByUserIDNull.Valid {
		transfer.CancelledByUserID = &cancelledByUserIDNull.Int64
	}
	if approvedAtNull.Valid {
		transfer.ApprovedAt = &approvedAtNull.Time
	}
	if dispatchedAtNull.Valid {
		transfer.DispatchedAt = &dispatchedAtNull.Time
	}
	if receivedAtNull.Valid {
		transfer.ReceivedAt = &receivedAtNull.Time
	}
	if cancelledAtNull.Valid {
		transfer.CancelledAt = &cancelledAtNull.Time
	}
	return transfer, nil
}

type transferQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func attachTransferItems(ctx context.Context, q transferQueryer, transfers []domaintransfer.Transfer) ([]domaintransfer.Transfer, error) {
	if len(transfers) == 0 {
		return transfers, nil
	}

	transferIndex := make(map[int64]int, len(transfers))
	transferIDs := make([]int64, 0, len(transfers))
	for index, transfer := range transfers {
		transferIndex[transfer.ID] = index
		transferIDs = append(transferIDs, transfer.ID)
	}

	rows, err := q.QueryContext(
		ctx,
		`SELECT iti.transfer_id, iti.product_id, p.sku, p.name, iti.quantity
		FROM inventory_transfer_items iti
		INNER JOIN products p ON p.id = iti.product_id
		WHERE iti.transfer_id = ANY($1)
		ORDER BY iti.transfer_id, iti.id`,
		pq.Array(transferIDs),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			transferID int64
			item       domaintransfer.Item
		)
		if err := rows.Scan(&transferID, &item.ProductID, &item.ProductSKU, &item.ProductName, &item.Quantity); err != nil {
			return nil, err
		}
		index := transferIndex[transferID]
		transfers[index].Items = append(transfers[index].Items, item)
	}

	return transfers, rows.Err()
}
