package postgrestransfer

import (
	"context"
	"database/sql"
	"errors"
	"sort"

	"github.com/lib/pq"

	transferapp "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer"
	domaintransfer "github.com/IanStuardo-Dev/backend-crud/internal/domain/transfer"
	domainuser "github.com/IanStuardo-Dev/backend-crud/internal/domain/user"
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
	if err = ensureTransferRequesterCanAct(ctx, tx, transfer.CompanyID, transfer.RequestedByUserID); err != nil {
		return err
	}
	if err = ensureTransferSupervisorEligible(ctx, tx, transfer.CompanyID, transfer.SupervisorUserID); err != nil {
		return err
	}

	productRows, err := loadBranchInventoryForTransfer(ctx, tx, transfer.CompanyID, transfer.OriginBranchID, transfer.Items, false)
	if err != nil {
		return err
	}

	transfer.Status = domaintransfer.StatusPendingApproval
	err = tx.QueryRowContext(
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
	if err != nil {
		if isForeignKeyViolation(err) {
			return transferapp.ErrInvalidReference
		}
		return err
	}

	for index, item := range transfer.Items {
		productRow, ok := productRows[item.ProductID]
		if !ok || productRow.ProductSKU == "" {
			return transferapp.ErrInvalidReference
		}

		transfer.Items[index].ProductSKU = productRow.ProductSKU
		transfer.Items[index].ProductName = productRow.ProductName

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
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *repository) Approve(ctx context.Context, transfer *domaintransfer.Transfer) error {
	_, err := r.db.ExecContext(
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

func (r *repository) Dispatch(ctx context.Context, transfer *domaintransfer.Transfer) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	productRows, err := loadBranchInventoryForTransfer(ctx, tx, transfer.CompanyID, transfer.OriginBranchID, transfer.Items, true)
	if err != nil {
		return err
	}

	for _, item := range transfer.Items {
		originRow, ok := productRows[item.ProductID]
		if !ok || originRow.ProductSKU == "" {
			return transferapp.ErrInvalidReference
		}
		if originRow.AvailableStock < item.Quantity {
			return transferapp.ErrInsufficientStock
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
			valueOrZero(transfer.DispatchedByUserID),
		); err != nil {
			return err
		}
	}

	if _, err = tx.ExecContext(
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
	); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *repository) Receive(ctx context.Context, transfer *domaintransfer.Transfer) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	destinationRows, err := loadBranchInventoryForTransfer(ctx, tx, transfer.CompanyID, transfer.DestinationBranchID, transfer.Items, true)
	if err != nil {
		return err
	}

	for _, item := range transfer.Items {
		destinationRow := destinationRows[item.ProductID]
		destinationStockAfter := destinationRow.StockOnHand + item.Quantity
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
			) VALUES ($1,$2,$3,NULL,$4,'transfer_in',$5,$6,$7,NOW())`,
			transfer.CompanyID,
			transfer.DestinationBranchID,
			item.ProductID,
			transfer.ID,
			item.Quantity,
			destinationStockAfter,
			valueOrZero(transfer.ReceivedByUserID),
		); err != nil {
			return err
		}
	}

	if _, err = tx.ExecContext(
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
	); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *repository) Cancel(ctx context.Context, transfer *domaintransfer.Transfer) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if transfer.DispatchedAt != nil && transfer.ReceivedAt == nil {
		originRows, err := loadBranchInventoryForTransfer(ctx, tx, transfer.CompanyID, transfer.OriginBranchID, transfer.Items, true)
		if err != nil {
			return err
		}

		for _, item := range transfer.Items {
			originRow, ok := originRows[item.ProductID]
			if !ok || originRow.ProductSKU == "" {
				return transferapp.ErrInvalidReference
			}

			originStockAfter := originRow.StockOnHand + item.Quantity
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
				) VALUES ($1,$2,$3,NULL,$4,'transfer_return',$5,$6,$7,NOW())`,
				transfer.CompanyID,
				transfer.OriginBranchID,
				item.ProductID,
				transfer.ID,
				item.Quantity,
				originStockAfter,
				valueOrZero(transfer.CancelledByUserID),
			); err != nil {
				return err
			}
		}
	}

	if _, err = tx.ExecContext(
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
	); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *repository) List(ctx context.Context) ([]domaintransfer.Transfer, error) {
	return r.listByQuery(ctx, `SELECT id, company_id, origin_branch_id, destination_branch_id, status, requested_by_user_id, supervisor_user_id,
			approved_by_user_id, dispatched_by_user_id, received_by_user_id, cancelled_by_user_id,
			note, created_at, approved_at, dispatched_at, received_at, cancelled_at
		FROM inventory_transfers
		ORDER BY id`)
}

func (r *repository) ListByBranch(ctx context.Context, branchID int64) ([]domaintransfer.Transfer, error) {
	return r.listByQuery(ctx, `SELECT id, company_id, origin_branch_id, destination_branch_id, status, requested_by_user_id, supervisor_user_id,
			approved_by_user_id, dispatched_by_user_id, received_by_user_id, cancelled_by_user_id,
			note, created_at, approved_at, dispatched_at, received_at, cancelled_at
		FROM inventory_transfers
		WHERE origin_branch_id = $1 OR destination_branch_id = $1
		ORDER BY id`, branchID)
}

func (r *repository) listByQuery(ctx context.Context, query string, args ...any) ([]domaintransfer.Transfer, error) {
	rows, err := r.db.QueryContext(
		ctx,
		query,
		args...,
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

func loadBranchInventoryForTransfer(ctx context.Context, tx *sql.Tx, companyID, branchID int64, items []domaintransfer.Item, forUpdate bool) (map[int64]transferInventoryRow, error) {
	productIDs := make([]int64, 0, len(items))
	for _, item := range items {
		productIDs = append(productIDs, item.ProductID)
	}
	sort.Slice(productIDs, func(i, j int) bool { return productIDs[i] < productIDs[j] })

	query := `SELECT bi.product_id, p.sku, p.name, bi.stock_on_hand, bi.reserved_stock, (bi.stock_on_hand - bi.reserved_stock) AS available_stock
		FROM branch_inventory bi
		INNER JOIN products p ON p.id = bi.product_id AND p.company_id = bi.company_id
		WHERE bi.company_id = $1 AND bi.branch_id = $2 AND bi.product_id = ANY($3)`
	if forUpdate {
		query += "\n\t\tFOR UPDATE"
	}

	rows, err := tx.QueryContext(ctx, query, companyID, branchID, pq.Array(productIDs))
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

func ensureTransferRequesterCanAct(ctx context.Context, tx *sql.Tx, companyID, userID int64) error {
	userRecord, err := loadTransferUser(ctx, tx, userID)
	if err != nil {
		return err
	}
	if !userRecord.IsActive {
		return transferapp.ErrInvalidReference
	}
	if userRecord.Role == domainuser.RoleSuperAdmin {
		return nil
	}
	if userRecord.CompanyID == nil || *userRecord.CompanyID != companyID {
		return transferapp.ErrInvalidReference
	}

	return nil
}

func ensureTransferSupervisorEligible(ctx context.Context, tx *sql.Tx, companyID, userID int64) error {
	userRecord, err := loadTransferUser(ctx, tx, userID)
	if err != nil {
		return err
	}
	if !userRecord.IsActive {
		return transferapp.ErrInvalidReference
	}
	if userRecord.Role == domainuser.RoleSuperAdmin {
		return nil
	}
	if userRecord.CompanyID == nil || *userRecord.CompanyID != companyID {
		return transferapp.ErrInvalidReference
	}
	if userRecord.Role != domainuser.RoleCompanyAdmin && userRecord.Role != domainuser.RoleInventoryManager {
		return transferapp.ErrInvalidReference
	}

	return nil
}

type transferUserRecord struct {
	CompanyID *int64
	Role      string
	IsActive  bool
}

func loadTransferUser(ctx context.Context, tx *sql.Tx, userID int64) (transferUserRecord, error) {
	var (
		record        transferUserRecord
		companyIDNull sql.NullInt64
	)

	err := tx.QueryRowContext(
		ctx,
		`SELECT company_id, role, is_active
		FROM users
		WHERE id = $1`,
		userID,
	).Scan(&companyIDNull, &record.Role, &record.IsActive)
	if errors.Is(err, sql.ErrNoRows) {
		return transferUserRecord{}, transferapp.ErrInvalidReference
	}
	if err != nil {
		return transferUserRecord{}, err
	}
	if companyIDNull.Valid {
		record.CompanyID = &companyIDNull.Int64
	}

	return record, nil
}

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
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transfers, nil
}

func isForeignKeyViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23503"
}

func valueOrZero(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}
