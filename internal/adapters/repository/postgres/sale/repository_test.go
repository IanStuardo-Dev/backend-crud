package postgressale

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domainsale "github.com/IanStuardo-Dev/backend-crud/internal/domain/sale"
)

func TestSaleStoreCreatePersistsSaleAndItems(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	store := NewSaleStore(db)
	now := time.Now().UTC()
	sale := &domainsale.Sale{
		CompanyID:        1,
		BranchID:         1,
		CreatedByUserID:  3,
		TotalAmountCents: 12000,
		Items: []domainsale.Item{{
			ProductID:      9,
			ProductSKU:     "SKU-009",
			ProductName:    "Monitor",
			Quantity:       2,
			UnitPriceCents: 6000,
			SubtotalCents:  12000,
		}},
	}

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO sales (company_id, branch_id, created_by_user_id, total_amount_cents, created_at)
		VALUES ($1,$2,$3,$4,NOW())
		RETURNING id, created_at`)).
		WithArgs(int64(1), int64(1), int64(3), int64(12000)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(7, now))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO sale_items (sale_id, product_id, product_sku, product_name, quantity, unit_price_cents, subtotal_cents)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`)).
		WithArgs(int64(7), int64(9), "SKU-009", "Monitor", 2, int64(6000), int64(12000)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := store.Create(context.Background(), sale); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if sale.ID != 7 || sale.CreatedAt != now {
		t.Fatalf("unexpected sale %#v", sale)
	}

	assertMockExpectations(t, mock)
}

func TestSaleStoreGetByIDReturnsNilWhenMissing(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	store := NewSaleStore(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, company_id, branch_id, created_by_user_id, total_amount_cents, created_at
		FROM sales
		WHERE id = $1`)).
		WithArgs(int64(10)).
		WillReturnError(sql.ErrNoRows)

	sale, err := store.GetByID(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if sale != nil {
		t.Fatalf("expected nil sale, got %#v", sale)
	}

	assertMockExpectations(t, mock)
}

func TestStockStoreLoadForSaleReturnsSnapshots(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	store := NewStockStore(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT p.id, p.sku, p.name, p.price_cents, bi.stock_on_hand, bi.reserved_stock
		FROM branch_inventory bi
		INNER JOIN products p ON p.id = bi.product_id AND p.company_id = bi.company_id
		WHERE bi.company_id = $1 AND bi.branch_id = $2 AND bi.product_id = ANY($3)
		FOR UPDATE`)).
		WithArgs(int64(1), int64(1), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "sku", "name", "price_cents", "stock_on_hand", "reserved_stock"}).
			AddRow(9, "SKU-009", "Monitor", 6000, 5, 1))

	snapshots, err := store.LoadForSale(context.Background(), 1, 1, []domainsale.Item{{ProductID: 9, Quantity: 2}}, true)
	if err != nil {
		t.Fatalf("LoadForSale() error = %v", err)
	}
	if snapshots[9].AvailableStock != 4 {
		t.Fatalf("unexpected snapshot %#v", snapshots[9])
	}

	assertMockExpectations(t, mock)
}

func TestReferenceStoreUserExistsReturnsTrue(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	store := NewReferenceStore(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE id = $1
		)`)).
		WithArgs(int64(3)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := store.UserExists(context.Background(), 3)
	if err != nil {
		t.Fatalf("UserExists() error = %v", err)
	}
	if !exists {
		t.Fatal("expected user to exist")
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
