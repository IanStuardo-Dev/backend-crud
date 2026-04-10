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

func TestTransferStoreCreateStoresPendingApprovalTransfer(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	store := NewTransferStore(db)
	now := time.Now().UTC()
	transfer := &domaintransfer.Transfer{
		CompanyID:           1,
		OriginBranchID:      1,
		DestinationBranchID: 2,
		RequestedByUserID:   7,
		SupervisorUserID:    9,
		Status:              domaintransfer.StatusPendingApproval,
		Note:                "Move fast sellers",
		Items: []domaintransfer.Item{{
			ProductID: 10,
			Quantity:  3,
		}},
	}

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

	if err := store.Create(context.Background(), transfer); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if transfer.ID != 4 || transfer.CreatedAt != now {
		t.Fatalf("unexpected transfer %#v", transfer)
	}

	assertMockExpectations(t, mock)
}

func TestTransferStoreGetByIDReturnsNilWhenMissing(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	store := NewTransferStore(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, company_id, origin_branch_id, destination_branch_id, status, requested_by_user_id, supervisor_user_id,
			approved_by_user_id, dispatched_by_user_id, received_by_user_id, cancelled_by_user_id,
			note, created_at, approved_at, dispatched_at, received_at, cancelled_at
		FROM inventory_transfers
		WHERE id = $1`)).
		WithArgs(int64(10)).
		WillReturnError(sql.ErrNoRows)

	transfer, err := store.GetByID(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if transfer != nil {
		t.Fatalf("expected nil transfer, got %#v", transfer)
	}

	assertMockExpectations(t, mock)
}

func TestTransferStockStoreLoadsSnapshots(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	store := NewStockStore(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT bi.product_id, p.sku, p.name, bi.stock_on_hand, bi.reserved_stock, (bi.stock_on_hand - bi.reserved_stock) AS available_stock
		FROM branch_inventory bi
		INNER JOIN products p ON p.id = bi.product_id AND p.company_id = bi.company_id
		WHERE bi.company_id = $1 AND bi.branch_id = $2 AND bi.product_id = ANY($3)
		FOR UPDATE`)).
		WithArgs(int64(1), int64(1), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"product_id", "sku", "name", "stock_on_hand", "reserved_stock", "available_stock"}).
			AddRow(10, "SKU-010", "Mechanical Keyboard", 8, 1, 7))

	rows, err := store.LoadForTransfer(context.Background(), 1, 1, []domaintransfer.Item{{ProductID: 10, Quantity: 2}}, true)
	if err != nil {
		t.Fatalf("LoadForTransfer() error = %v", err)
	}
	if rows[10].AvailableStock != 7 {
		t.Fatalf("unexpected stock row %#v", rows[10])
	}

	assertMockExpectations(t, mock)
}

func TestTransferReferenceStoreSupervisorEligible(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	store := NewReferenceStore(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT company_id, role, is_active
		FROM users
		WHERE id = $1`)).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"company_id", "role", "is_active"}).AddRow(1, "company_admin", true))

	ok, err := store.SupervisorEligible(context.Background(), 1, 9)
	if err != nil {
		t.Fatalf("SupervisorEligible() error = %v", err)
	}
	if !ok {
		t.Fatal("expected supervisor to be eligible")
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
