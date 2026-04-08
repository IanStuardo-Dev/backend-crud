package postgresproduct

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	productapp "github.com/IanStuardo-Dev/backend-crud/internal/application/product"
	domainproduct "github.com/IanStuardo-Dev/backend-crud/internal/domain/product"
	"github.com/lib/pq"
)

func TestRepositoryCreateAssignsIDAndTimestamps(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	repo := NewRepository(db)
	product := &domainproduct.Product{
		CompanyID:  1,
		BranchID:   1,
		SKU:        "SKU-001",
		Name:       "Laptop Stand",
		Category:   "office",
		PriceCents: 3499,
		Currency:   "USD",
	}
	now := time.Now().UTC()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO products (
			company_id,
			branch_id,
			sku,
			name,
			description,
			category,
			brand,
			price_cents,
			currency,
			stock,
			embedding,
			created_at,
			updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::vector,NOW(),NOW())
		RETURNING id, created_at, updated_at`)).
		WithArgs(int64(1), int64(1), "SKU-001", "Laptop Stand", "", "office", "", int64(3499), "USD", 0, nil).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(5, now, now))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO branch_inventory (
			company_id,
			branch_id,
			product_id,
			stock_on_hand,
			reserved_stock,
			created_at,
			updated_at
		) VALUES ($1,$2,$3,$4,0,NOW(),NOW())`)).
		WithArgs(int64(1), int64(1), int64(5), 0).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repo.Create(context.Background(), product); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if product.ID != 5 {
		t.Fatalf("expected ID 5, got %d", product.ID)
	}
	if product.CreatedAt != now || product.UpdatedAt != now {
		t.Fatalf("expected timestamps to be populated")
	}

	assertMockExpectations(t, mock)
}

func TestRepositoryCreateReturnsConflictOnUniqueViolation(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	repo := NewRepository(db)
	product := &domainproduct.Product{
		CompanyID:  1,
		BranchID:   1,
		SKU:        "SKU-001",
		Name:       "Laptop Stand",
		Category:   "office",
		PriceCents: 3499,
		Currency:   "USD",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO products (
			company_id,
			branch_id,
			sku,
			name,
			description,
			category,
			brand,
			price_cents,
			currency,
			stock,
			embedding,
			created_at,
			updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::vector,NOW(),NOW())
		RETURNING id, created_at, updated_at`)).
		WithArgs(int64(1), int64(1), "SKU-001", "Laptop Stand", "", "office", "", int64(3499), "USD", 0, nil).
		WillReturnError(&pq.Error{Code: "23505"})
	mock.ExpectRollback()

	err := repo.Create(context.Background(), product)
	if !errors.Is(err, productapp.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}

	assertMockExpectations(t, mock)
}

func TestRepositoryListReturnsProducts(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	repo := NewRepository(db)
	now := time.Now().UTC()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT p.id, p.company_id, p.branch_id, p.sku, p.name, p.description, p.category, p.brand, p.price_cents, p.currency,
			COALESCE(SUM(bi.stock_on_hand), p.stock) AS stock,
			p.embedding::text, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN branch_inventory bi ON bi.product_id = p.id
		GROUP BY p.id
		ORDER BY p.id`)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "company_id", "branch_id", "sku", "name", "description", "category", "brand", "price_cents", "currency", "stock", "embedding", "created_at", "updated_at",
		}).
			AddRow(1, 1, 1, "SKU-001", "Laptop Stand", "Aluminum stand", "office", "Acme", 3499, "USD", 5, "[0.1,0.2]", now, now).
			AddRow(2, 1, 1, "SKU-002", "Desk Lamp", "Warm light", "office", "Acme", 2599, "USD", 3, nil, now, now))

	products, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(products) != 2 {
		t.Fatalf("expected 2 products, got %d", len(products))
	}
	if len(products[0].Embedding) != 2 || products[1].Embedding != nil {
		t.Fatalf("unexpected embeddings: %#v", products)
	}

	assertMockExpectations(t, mock)
}

func TestRepositoryGetByIDReturnsNilWhenMissing(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	repo := NewRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT p.id, p.company_id, p.branch_id, p.sku, p.name, p.description, p.category, p.brand, p.price_cents, p.currency,
			COALESCE(SUM(bi.stock_on_hand), p.stock) AS stock,
			p.embedding::text, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN branch_inventory bi ON bi.product_id = p.id
		WHERE p.id=$1
		GROUP BY p.id`)).
		WithArgs(int64(10)).
		WillReturnError(sql.ErrNoRows)

	product, err := repo.GetByID(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if product != nil {
		t.Fatalf("expected nil product, got %#v", product)
	}

	assertMockExpectations(t, mock)
}

func TestRepositoryFindNeighborsReturnsRankedNeighbors(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	repo := NewRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT
			p.id,
			p.sku,
			p.name,
			p.description,
			p.category,
			p.brand,
			p.price_cents,
			p.currency,
			(source.embedding <=> p.embedding) AS distance
		FROM products AS source
		JOIN products AS p
			ON p.company_id = source.company_id
			AND p.id <> source.id
			AND p.embedding IS NOT NULL
		WHERE source.id = $1
			AND source.company_id = $2
			AND source.embedding IS NOT NULL
			AND (1 - (source.embedding <=> p.embedding)) >= $3
		ORDER BY source.embedding <=> p.embedding ASC
		LIMIT $4`)).
		WithArgs(int64(4), int64(1), 0.35, 5).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "sku", "name", "description", "category", "brand", "price_cents", "currency", "distance",
		}).
			AddRow(13, "SEED-PRD-010", "Wireless Ergonomic Mouse", "Mouse ergonomico", "peripherals", "Acme", 25990, "CLP", 0.302876).
			AddRow(8, "SEED-PRD-005", "Laptop Stand Aluminum", "Soporte", "office", "Acme", 24990, "CLP", 0.643743))

	neighbors, err := repo.FindNeighbors(context.Background(), 4, 1, 5, 0.35)
	if err != nil {
		t.Fatalf("FindNeighbors() error = %v", err)
	}
	if len(neighbors) != 2 {
		t.Fatalf("expected 2 neighbors, got %d", len(neighbors))
	}
	if neighbors[0].ProductID != 13 || neighbors[0].Distance != 0.302876 {
		t.Fatalf("unexpected first neighbor %#v", neighbors[0])
	}

	assertMockExpectations(t, mock)
}

func TestRepositoryUpdateReturnsNotFoundWhenMissing(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	repo := NewRepository(db)
	product := &domainproduct.Product{
		ID:         7,
		CompanyID:  1,
		BranchID:   1,
		SKU:        "SKU-007",
		Name:       "Headphones",
		Category:   "audio",
		PriceCents: 15999,
		Currency:   "USD",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`UPDATE products
		SET company_id=$1, branch_id=$2, sku=$3, name=$4, description=$5, category=$6, brand=$7, price_cents=$8, currency=$9, stock=$10, embedding=$11::vector, updated_at=NOW()
		WHERE id=$12
		RETURNING created_at, updated_at`)).
		WithArgs(int64(1), int64(1), "SKU-007", "Headphones", "", "audio", "", int64(15999), "USD", 0, nil, int64(7)).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	err := repo.Update(context.Background(), product)
	if !errors.Is(err, productapp.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	assertMockExpectations(t, mock)
}

func TestRepositoryDeleteReturnsNotFoundWhenNoRowsAffected(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	repo := NewRepository(db)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM products WHERE id=$1")).
		WithArgs(int64(8)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), 8)
	if !errors.Is(err, productapp.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
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
