package postgrestransfer

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	domaintransfer "github.com/IanStuardo-Dev/backend-crud/internal/domain/transfer"
)

func TestRepositoryCreateStoresPendingApprovalTransfer(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	repo := NewRepository(db)
	transfer := &domaintransfer.Transfer{
		CompanyID:           1,
		OriginBranchID:      1,
		DestinationBranchID: 2,
		RequestedByUserID:   7,
		SupervisorUserID:    9,
		Note:                "Move fast sellers",
		Items: []domaintransfer.Item{{
			ProductID: 10,
			Quantity:  3,
		}},
	}
	now := time.Now().UTC()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS (
			SELECT 1
			FROM branches
			WHERE company_id = $1 AND id = $2
		)`)).
		WithArgs(int64(1), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS (
			SELECT 1
			FROM branches
			WHERE company_id = $1 AND id = $2
		)`)).
		WithArgs(int64(1), int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT company_id, role, is_active
		FROM users
		WHERE id = $1`)).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"company_id", "role", "is_active"}).AddRow(1, "inventory_manager", true))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT company_id, role, is_active
		FROM users
		WHERE id = $1`)).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"company_id", "role", "is_active"}).AddRow(1, "company_admin", true))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT bi.product_id, p.sku, p.name, bi.stock_on_hand, bi.reserved_stock, (bi.stock_on_hand - bi.reserved_stock) AS available_stock
		FROM branch_inventory bi
		INNER JOIN products p ON p.id = bi.product_id AND p.company_id = bi.company_id
		WHERE bi.company_id = $1 AND bi.branch_id = $2 AND bi.product_id = ANY($3)`)).
		WithArgs(int64(1), int64(1), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"product_id", "sku", "name", "stock_on_hand", "reserved_stock", "available_stock"}).
			AddRow(10, "SKU-010", "Mechanical Keyboard", 8, 0, 8))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO inventory_transfers (
			company_id,
			origin_branch_id,
			destination_branch_id,
			status,
			requested_by_user_id,
			supervisor_user_id,
			note,
			created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,NOW())
		RETURNING id, created_at`)).
		WithArgs(int64(1), int64(1), int64(2), domaintransfer.StatusPendingApproval, int64(7), int64(9), "Move fast sellers").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(4, now))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO inventory_transfer_items (transfer_id, product_id, quantity)
			VALUES ($1,$2,$3)`)).
		WithArgs(int64(4), int64(10), 3).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repo.Create(context.Background(), transfer); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if transfer.ID != 4 || transfer.Status != domaintransfer.StatusPendingApproval {
		t.Fatalf("unexpected transfer: %#v", transfer)
	}
	assertMockExpectations(t, mock)
}

func TestRepositoryDispatchMovesStockOut(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	repo := NewRepository(db)
	now := time.Now().UTC()
	actorID := int64(8)
	transfer := &domaintransfer.Transfer{
		ID:                  11,
		CompanyID:           1,
		OriginBranchID:      1,
		DestinationBranchID: 2,
		Status:              domaintransfer.StatusInTransit,
		RequestedByUserID:   7,
		SupervisorUserID:    9,
		DispatchedByUserID:  &actorID,
		DispatchedAt:        &now,
		Items: []domaintransfer.Item{{
			ProductID: 10,
			Quantity:  2,
		}},
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT bi.product_id, p.sku, p.name, bi.stock_on_hand, bi.reserved_stock, (bi.stock_on_hand - bi.reserved_stock) AS available_stock
		FROM branch_inventory bi
		INNER JOIN products p ON p.id = bi.product_id AND p.company_id = bi.company_id
		WHERE bi.company_id = $1 AND bi.branch_id = $2 AND bi.product_id = ANY($3)
		FOR UPDATE`)).
		WithArgs(int64(1), int64(1), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"product_id", "sku", "name", "stock_on_hand", "reserved_stock", "available_stock"}).
			AddRow(10, "SKU-010", "Mechanical Keyboard", 8, 1, 7))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE branch_inventory
			SET stock_on_hand = $1, updated_at = NOW()
			WHERE company_id = $2 AND branch_id = $3 AND product_id = $4`)).
		WithArgs(6, int64(1), int64(1), int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO inventory_movements (
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
			) VALUES ($1,$2,$3,NULL,$4,'transfer_out',$5,$6,$7,NOW())`)).
		WithArgs(int64(1), int64(1), int64(10), int64(11), -2, 6, int64(8)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE inventory_transfers
		SET status = $1,
			dispatched_by_user_id = $2,
			dispatched_at = $3
		WHERE id = $4`)).
		WithArgs(domaintransfer.StatusInTransit, &actorID, transfer.DispatchedAt, int64(11)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.Dispatch(context.Background(), transfer); err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}
	assertMockExpectations(t, mock)
}

func TestRepositoryCancelReturnsStockWhenTransferWasInTransit(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	repo := NewRepository(db)
	dispatchedAt := time.Now().UTC().Add(-time.Hour)
	cancelledAt := time.Now().UTC()
	actorID := int64(12)
	transfer := &domaintransfer.Transfer{
		ID:                  15,
		CompanyID:           1,
		OriginBranchID:      1,
		DestinationBranchID: 2,
		Status:              domaintransfer.StatusCancelled,
		RequestedByUserID:   7,
		SupervisorUserID:    9,
		DispatchedAt:        &dispatchedAt,
		CancelledByUserID:   &actorID,
		CancelledAt:         &cancelledAt,
		Items: []domaintransfer.Item{{
			ProductID: 10,
			Quantity:  2,
		}},
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT bi.product_id, p.sku, p.name, bi.stock_on_hand, bi.reserved_stock, (bi.stock_on_hand - bi.reserved_stock) AS available_stock
		FROM branch_inventory bi
		INNER JOIN products p ON p.id = bi.product_id AND p.company_id = bi.company_id
		WHERE bi.company_id = $1 AND bi.branch_id = $2 AND bi.product_id = ANY($3)
		FOR UPDATE`)).
		WithArgs(int64(1), int64(1), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"product_id", "sku", "name", "stock_on_hand", "reserved_stock", "available_stock"}).
			AddRow(10, "SKU-010", "Mechanical Keyboard", 6, 0, 6))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE branch_inventory
				SET stock_on_hand = $1, updated_at = NOW()
				WHERE company_id = $2 AND branch_id = $3 AND product_id = $4`)).
		WithArgs(8, int64(1), int64(1), int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO inventory_movements (
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
				) VALUES ($1,$2,$3,NULL,$4,'transfer_return',$5,$6,$7,NOW())`)).
		WithArgs(int64(1), int64(1), int64(10), int64(15), 2, 8, int64(12)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE inventory_transfers
		SET status = $1,
			cancelled_by_user_id = $2,
			cancelled_at = $3
		WHERE id = $4`)).
		WithArgs(domaintransfer.StatusCancelled, &actorID, transfer.CancelledAt, int64(15)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.Cancel(context.Background(), transfer); err != nil {
		t.Fatalf("Cancel() error = %v", err)
	}
	assertMockExpectations(t, mock)
}

func TestRepositoryListByBranchQueriesOriginAndDestination(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	repo := NewRepository(db)
	now := time.Now().UTC()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, company_id, origin_branch_id, destination_branch_id, status, requested_by_user_id, supervisor_user_id,
			approved_by_user_id, dispatched_by_user_id, received_by_user_id, cancelled_by_user_id,
			note, created_at, approved_at, dispatched_at, received_at, cancelled_at
		FROM inventory_transfers
		WHERE origin_branch_id = $1 OR destination_branch_id = $1
		ORDER BY id`)).
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "company_id", "origin_branch_id", "destination_branch_id", "status", "requested_by_user_id", "supervisor_user_id",
			"approved_by_user_id", "dispatched_by_user_id", "received_by_user_id", "cancelled_by_user_id",
			"note", "created_at", "approved_at", "dispatched_at", "received_at", "cancelled_at",
		}).AddRow(3, 1, 1, 2, domaintransfer.StatusInTransit, 7, 9, 9, 8, nil, nil, "restock", now, now, now, nil, nil))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT iti.transfer_id, iti.product_id, p.sku, p.name, iti.quantity
		FROM inventory_transfer_items iti
		INNER JOIN products p ON p.id = iti.product_id
		WHERE iti.transfer_id = ANY($1)
		ORDER BY iti.transfer_id, iti.id`)).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"transfer_id", "product_id", "sku", "name", "quantity"}).
			AddRow(3, 10, "SKU-010", "Mechanical Keyboard", 2))

	transfers, err := repo.ListByBranch(context.Background(), 2)
	if err != nil {
		t.Fatalf("ListByBranch() error = %v", err)
	}
	if len(transfers) != 1 || transfers[0].DestinationBranchID != 2 {
		t.Fatalf("unexpected transfers: %#v", transfers)
	}
	assertMockExpectations(t, mock)
}

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	return db, mock
}

func assertMockExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet mock expectations: %v", err)
	}
}
