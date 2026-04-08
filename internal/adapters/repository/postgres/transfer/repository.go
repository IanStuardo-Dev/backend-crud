package postgrestransfer

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"time"

	"github.com/lib/pq"

	transferapp "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer"
	domaintransfer "github.com/IanStuardo-Dev/backend-crud/internal/domain/transfer"
)

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) transferapp.Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, transfer *domaintransfer.Transfer) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = ensureTransferBranchExists(ctx, tx, transfer.CompanyID, transfer.OriginBranchID); err != nil {
		return err
	}
	if err = ensureTransferBranchExists(ctx, tx, transfer.CompanyID, transfer.DestinationBranchID); err != nil {
		return err
	}
	if err = ensureTransferUserExists(ctx, tx, transfer.RequestedByUserID); err != nil {
		return err
	}

	productRows, err := loadOriginInventoryForTransfer(ctx, tx, transfer.CompanyID, transfer.OriginBranchID, transfer.Items)
	if err != nil {
		return err
	}
	destinationRows, err := loadDestinationInventoryForTransfer(ctx, tx, transfer.CompanyID, transfer.DestinationBranchID, transfer.Items)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	transfer.Status = "completed"
	transfer.CompletedByUserID = transfer.RequestedByUserID
	transfer.CompletedAt = &now

	err = tx.QueryRowContext(
		ctx,
		`INSERT INTO inventory_transfers (
			company_id,
			origin_branch_id,
			destination_branch_id,
			status,
			requested_by_user_id,
			completed_by_user_id,
			note,
			created_at,
			completed_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,NOW(),NOW())
		RETURNING id, created_at, completed_at`,
		transfer.CompanyID,
		transfer.OriginBranchID,
		transfer.DestinationBranchID,
		transfer.Status,
		transfer.RequestedByUserID,
		transfer.CompletedByUserID,
		transfer.Note,
	).Scan(&transfer.ID, &transfer.CreatedAt, &transfer.CompletedAt)
	if err != nil {
		if isForeignKeyViolation(err) {
			return transferapp.ErrInvalidReference
		}
		return err
	}

	for index, item := range transfer.Items {
		originRow, ok := productRows[item.ProductID]
		if !ok {
			return transferapp.ErrInvalidReference
		}
		if originRow.AvailableStock < item.Quantity {
			return transferapp.ErrInsufficientStock
		}

		transfer.Items[index].ProductSKU = originRow.ProductSKU
		transfer.Items[index].ProductName = originRow.ProductName

		if _, err = tx.ExecContext(
			ctx,
			`INSERT INTO inventory_transfer_items (transfer_id, product_id, quantity)
			VALUES ($1,$2,$3)`,
			transfer.ID,
			item.ProductID,
			item.Quantity,
		); err != nil {
			return err
		}

		originStockAfter := originRow.StockOnHand - item.Quantity
		if _, err = tx.ExecContext(
			ctx,
			`UPDATE branch_inventory
			SET stock_on_hand = $1, updated_at = NOW()
			WHERE company_id = $2 AND branch_id = $3 AND product_id = $4`,
			originStockAfter,
			transfer.CompanyID,
			transfer.OriginBranchID,
			item.ProductID,
		); err != nil {
			return err
		}

		destinationStockAfter := destinationRows[item.ProductID].StockOnHand + item.Quantity
		if _, err = tx.ExecContext(
			ctx,
			`INSERT INTO branch_inventory (
				company_id,
				branch_id,
				product_id,
				stock_on_hand,
				reserved_stock,
				created_at,
				updated_at
			) VALUES ($1,$2,$3,$4,0,NOW(),NOW())
			ON CONFLICT (branch_id, product_id)
			DO UPDATE SET
				company_id = EXCLUDED.company_id,
				stock_on_hand = EXCLUDED.stock_on_hand,
				updated_at = NOW()`,
			transfer.CompanyID,
			transfer.DestinationBranchID,
			item.ProductID,
			destinationStockAfter,
		); err != nil {
			return err
		}

		if _, err = tx.ExecContext(
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
			) VALUES ($1,$2,$3,NULL,$4,'transfer_out',$5,$6,$7,NOW())`,
			transfer.CompanyID,
			transfer.OriginBranchID,
			item.ProductID,
			transfer.ID,
			-item.Quantity,
			originStockAfter,
			transfer.RequestedByUserID,
		); err != nil {
			return err
		}

		if _, err = tx.ExecContext(
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
			) VALUES ($1,$2,$3,NULL,$4,'transfer_in',$5,$6,$7,NOW())`,
			transfer.CompanyID,
			transfer.DestinationBranchID,
			item.ProductID,
			transfer.ID,
			item.Quantity,
			destinationStockAfter,
			transfer.RequestedByUserID,
		); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *repository) List(ctx context.Context) ([]domaintransfer.Transfer, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, company_id, origin_branch_id, destination_branch_id, status, requested_by_user_id, completed_by_user_id, note, created_at, completed_at
		FROM inventory_transfers
		ORDER BY id`,
	)
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

	return attachTransferItems(ctx, r.db, transfers)
}

func (r *repository) GetByID(ctx context.Context, id int64) (*domaintransfer.Transfer, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, company_id, origin_branch_id, destination_branch_id, status, requested_by_user_id, completed_by_user_id, note, created_at, completed_at
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

	transfers, err := attachTransferItems(ctx, r.db, []domaintransfer.Transfer{transfer})
	if err != nil {
		return nil, err
	}

	return &transfers[0], nil
}

type transferInventoryRow struct {
	ProductID      int64
	ProductSKU     string
	ProductName    string
	StockOnHand    int
	ReservedStock  int
	AvailableStock int
}

func loadOriginInventoryForTransfer(ctx context.Context, tx *sql.Tx, companyID, branchID int64, items []domaintransfer.Item) (map[int64]transferInventoryRow, error) {
	productIDs := make([]int64, 0, len(items))
	for _, item := range items {
		productIDs = append(productIDs, item.ProductID)
	}
	sort.Slice(productIDs, func(i, j int) bool { return productIDs[i] < productIDs[j] })

	rows, err := tx.QueryContext(
		ctx,
		`SELECT bi.product_id, p.sku, p.name, bi.stock_on_hand, bi.reserved_stock, (bi.stock_on_hand - bi.reserved_stock) AS available_stock
		FROM branch_inventory bi
		INNER JOIN products p ON p.id = bi.product_id AND p.company_id = bi.company_id
		WHERE bi.company_id = $1 AND bi.branch_id = $2 AND bi.product_id = ANY($3)
		FOR UPDATE`,
		companyID,
		branchID,
		pq.Array(productIDs),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	inventoryRows := make(map[int64]transferInventoryRow, len(items))
	for rows.Next() {
		var row transferInventoryRow
		if err := rows.Scan(&row.ProductID, &row.ProductSKU, &row.ProductName, &row.StockOnHand, &row.ReservedStock, &row.AvailableStock); err != nil {
			return nil, err
		}
		inventoryRows[row.ProductID] = row
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return inventoryRows, nil
}

func loadDestinationInventoryForTransfer(ctx context.Context, tx *sql.Tx, companyID, branchID int64, items []domaintransfer.Item) (map[int64]transferInventoryRow, error) {
	productIDs := make([]int64, 0, len(items))
	for _, item := range items {
		productIDs = append(productIDs, item.ProductID)
	}
	sort.Slice(productIDs, func(i, j int) bool { return productIDs[i] < productIDs[j] })

	rows, err := tx.QueryContext(
		ctx,
		`SELECT bi.product_id, p.sku, p.name, bi.stock_on_hand, bi.reserved_stock, (bi.stock_on_hand - bi.reserved_stock) AS available_stock
		FROM branch_inventory bi
		INNER JOIN products p ON p.id = bi.product_id AND p.company_id = bi.company_id
		WHERE bi.company_id = $1 AND bi.branch_id = $2 AND bi.product_id = ANY($3)
		FOR UPDATE`,
		companyID,
		branchID,
		pq.Array(productIDs),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	inventoryRows := make(map[int64]transferInventoryRow, len(items))
	for rows.Next() {
		var row transferInventoryRow
		if err := rows.Scan(&row.ProductID, &row.ProductSKU, &row.ProductName, &row.StockOnHand, &row.ReservedStock, &row.AvailableStock); err != nil {
			return nil, err
		}
		inventoryRows[row.ProductID] = row
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, item := range items {
		if _, ok := inventoryRows[item.ProductID]; !ok {
			inventoryRows[item.ProductID] = transferInventoryRow{ProductID: item.ProductID}
		}
	}

	return inventoryRows, nil
}

func ensureTransferBranchExists(ctx context.Context, tx *sql.Tx, companyID, branchID int64) error {
	var exists bool
	err := tx.QueryRowContext(
		ctx,
		`SELECT EXISTS (
			SELECT 1
			FROM branches
			WHERE company_id = $1 AND id = $2
		)`,
		companyID,
		branchID,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return transferapp.ErrInvalidReference
	}
	return nil
}

func ensureTransferUserExists(ctx context.Context, tx *sql.Tx, userID int64) error {
	var exists bool
	err := tx.QueryRowContext(
		ctx,
		`SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE id = $1
		)`,
		userID,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return transferapp.ErrInvalidReference
	}
	return nil
}

type transferScanner interface {
	Scan(dest ...any) error
}

func scanTransfer(row transferScanner) (domaintransfer.Transfer, error) {
	var (
		transfer    domaintransfer.Transfer
		completedAt sql.NullTime
	)
	err := row.Scan(
		&transfer.ID,
		&transfer.CompanyID,
		&transfer.OriginBranchID,
		&transfer.DestinationBranchID,
		&transfer.Status,
		&transfer.RequestedByUserID,
		&transfer.CompletedByUserID,
		&transfer.Note,
		&transfer.CreatedAt,
		&completedAt,
	)
	if err != nil {
		return domaintransfer.Transfer{}, err
	}
	if completedAt.Valid {
		transfer.CompletedAt = &completedAt.Time
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
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transfers, nil
}

func isForeignKeyViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23503"
}
