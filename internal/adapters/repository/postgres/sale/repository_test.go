package postgressale

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	saleapp "github.com/IanStuardo-Dev/backend-crud/internal/application/sale"
	domainsale "github.com/IanStuardo-Dev/backend-crud/internal/domain/sale"
)

func TestRepositoryCreateDiscountsInventoryAndRegistersSale(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	repo := NewRepository(db)
	sale := &domainsale.Sale{
		CompanyID:       1,
		BranchID:        1,
		CreatedByUserID: 3,
		Items: []domainsale.Item{{
			ProductID: 9,
			Quantity:  2,
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
			FROM users
			WHERE id = $1
		)`)).
		WithArgs(int64(3)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT p.id, p.sku, p.name, p.price_cents, bi.stock_on_hand, bi.reserved_stock
		FROM branch_inventory bi
		INNER JOIN products p ON p.id = bi.product_id AND p.company_id = bi.company_id
		WHERE bi.company_id = $1 AND bi.branch_id = $2 AND bi.product_id = ANY($3)
		FOR UPDATE`)).
		WithArgs(int64(1), int64(1), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "sku", "name", "price_cents", "stock_on_hand", "reserved_stock"}).
			AddRow(9, "SKU-009", "Monitor", 6000, 5, 0))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO sales (company_id, branch_id, created_by_user_id, total_amount_cents, created_at)
		VALUES ($1,$2,$3,$4,NOW())
		RETURNING id, created_at`)).
		WithArgs(int64(1), int64(1), int64(3), int64(12000)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(7, now))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO sale_items (sale_id, product_id, product_sku, product_name, quantity, unit_price_cents, subtotal_cents)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`)).
		WithArgs(int64(7), int64(9), "SKU-009", "Monitor", 2, int64(6000), int64(12000)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE branch_inventory
			SET stock_on_hand = $1, updated_at = NOW()
			WHERE company_id = $2 AND branch_id = $3 AND product_id = $4`)).
		WithArgs(3, int64(1), int64(1), int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO inventory_movements (
				company_id,
				branch_id,
				product_id,
				sale_id,
				movement_type,
				quantity_delta,
				stock_after,
				created_by_user_id,
				created_at
			) VALUES ($1,$2,$3,$4,'sale',$5,$6,$7,NOW())`)).
		WithArgs(int64(1), int64(1), int64(9), int64(7), -2, 3, int64(3)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repo.Create(context.Background(), sale); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if sale.ID != 7 || sale.TotalAmountCents != 12000 {
		t.Fatalf("unexpected sale: %#v", sale)
	}
	assertMockExpectations(t, mock)
}

func TestRepositoryCreateReturnsInsufficientStock(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	repo := NewRepository(db)
	sale := &domainsale.Sale{
		CompanyID:       1,
		BranchID:        1,
		CreatedByUserID: 3,
		Items: []domainsale.Item{{
			ProductID: 9,
			Quantity:  10,
		}},
	}

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
			FROM users
			WHERE id = $1
		)`)).
		WithArgs(int64(3)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT p.id, p.sku, p.name, p.price_cents, bi.stock_on_hand, bi.reserved_stock
		FROM branch_inventory bi
		INNER JOIN products p ON p.id = bi.product_id AND p.company_id = bi.company_id
		WHERE bi.company_id = $1 AND bi.branch_id = $2 AND bi.product_id = ANY($3)
		FOR UPDATE`)).
		WithArgs(int64(1), int64(1), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "sku", "name", "price_cents", "stock_on_hand", "reserved_stock"}).
			AddRow(9, "SKU-009", "Monitor", 6000, 5, 0))
	mock.ExpectRollback()

	err := repo.Create(context.Background(), sale)
	if !errors.Is(err, saleapp.ErrInsufficientStock) {
		t.Fatalf("expected ErrInsufficientStock, got %v", err)
	}
	assertMockExpectations(t, mock)
}

func TestRepositoryGetByIDReturnsNilWhenMissing(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	repo := NewRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, company_id, branch_id, created_by_user_id, total_amount_cents, created_at
		FROM sales
		WHERE id = $1`)).
		WithArgs(int64(10)).
		WillReturnError(sql.ErrNoRows)

	sale, err := repo.GetByID(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if sale != nil {
		t.Fatalf("expected nil sale, got %#v", sale)
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
