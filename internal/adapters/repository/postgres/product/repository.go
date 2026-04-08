package postgresproduct

import (
	"context"
	"database/sql"
	"errors"

	productapp "github.com/IanStuardo-Dev/backend-crud/internal/application/product"
	domainproduct "github.com/IanStuardo-Dev/backend-crud/internal/domain/product"
	"github.com/lib/pq"
)

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) productapp.Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, product *domainproduct.Product) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	err = tx.QueryRowContext(
		ctx,
		`INSERT INTO products (
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
		RETURNING id, created_at, updated_at`,
		product.CompanyID,
		product.BranchID,
		product.SKU,
		product.Name,
		product.Description,
		product.Category,
		product.Brand,
		product.PriceCents,
		product.Currency,
		product.Stock,
		formatVector(product.Embedding),
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
	if isUniqueViolation(err) {
		return productapp.ErrConflict
	}
	if isForeignKeyViolation(err) {
		return productapp.ErrInvalidReference
	}
	if err != nil {
		return err
	}

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
		) VALUES ($1,$2,$3,$4,0,NOW(),NOW())`,
		product.CompanyID,
		product.BranchID,
		product.ID,
		product.Stock,
	); err != nil {
		if isUniqueViolation(err) {
			return productapp.ErrConflict
		}
		if isForeignKeyViolation(err) {
			return productapp.ErrInvalidReference
		}
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *repository) List(ctx context.Context) ([]domainproduct.Product, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT p.id, p.company_id, p.branch_id, p.sku, p.name, p.description, p.category, p.brand, p.price_cents, p.currency,
			COALESCE(SUM(bi.stock_on_hand), p.stock) AS stock,
			p.embedding::text, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN branch_inventory bi ON bi.product_id = p.id
		GROUP BY p.id
		ORDER BY p.id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]domainproduct.Product, 0)
	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func (r *repository) GetByID(ctx context.Context, id int64) (*domainproduct.Product, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT p.id, p.company_id, p.branch_id, p.sku, p.name, p.description, p.category, p.brand, p.price_cents, p.currency,
			COALESCE(SUM(bi.stock_on_hand), p.stock) AS stock,
			p.embedding::text, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN branch_inventory bi ON bi.product_id = p.id
		WHERE p.id=$1
		GROUP BY p.id`,
		id,
	)

	product, err := scanProduct(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *repository) FindNeighbors(ctx context.Context, sourceProductID, companyID int64, limit int, minSimilarity float64) ([]productapp.NeighborOutput, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT
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
		LIMIT $4`,
		sourceProductID,
		companyID,
		minSimilarity,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	neighbors := make([]productapp.NeighborOutput, 0)
	for rows.Next() {
		var neighbor productapp.NeighborOutput
		if err := rows.Scan(
			&neighbor.ProductID,
			&neighbor.SKU,
			&neighbor.Name,
			&neighbor.Description,
			&neighbor.Category,
			&neighbor.Brand,
			&neighbor.PriceCents,
			&neighbor.Currency,
			&neighbor.Distance,
		); err != nil {
			return nil, err
		}
		neighbors = append(neighbors, neighbor)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return neighbors, nil
}

func (r *repository) Update(ctx context.Context, product *domainproduct.Product) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	err = tx.QueryRowContext(
		ctx,
		`UPDATE products
		SET company_id=$1, branch_id=$2, sku=$3, name=$4, description=$5, category=$6, brand=$7, price_cents=$8, currency=$9, stock=$10, embedding=$11::vector, updated_at=NOW()
		WHERE id=$12
		RETURNING created_at, updated_at`,
		product.CompanyID,
		product.BranchID,
		product.SKU,
		product.Name,
		product.Description,
		product.Category,
		product.Brand,
		product.PriceCents,
		product.Currency,
		product.Stock,
		formatVector(product.Embedding),
		product.ID,
	).Scan(&product.CreatedAt, &product.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return productapp.ErrNotFound
	}
	if isUniqueViolation(err) {
		return productapp.ErrConflict
	}
	if isForeignKeyViolation(err) {
		return productapp.ErrInvalidReference
	}
	if err != nil {
		return err
	}

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
		product.CompanyID,
		product.BranchID,
		product.ID,
		product.Stock,
	); err != nil {
		if isUniqueViolation(err) {
			return productapp.ErrConflict
		}
		if isForeignKeyViolation(err) {
			return productapp.ErrInvalidReference
		}
		return err
	}

	if err = syncProductTotalStock(ctx, tx, []int64{product.ID}); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM products WHERE id=$1", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return productapp.ErrNotFound
	}

	return nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

func isForeignKeyViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23503"
}

type scanner interface {
	Scan(dest ...any) error
}

type execer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func syncProductTotalStock(ctx context.Context, exec execer, productIDs []int64) error {
	if len(productIDs) == 0 {
		return nil
	}

	_, err := exec.ExecContext(
		ctx,
		`UPDATE products p
		SET stock = totals.total_stock,
			updated_at = NOW()
		FROM (
			SELECT product_id, SUM(stock_on_hand) AS total_stock
			FROM branch_inventory
			WHERE product_id = ANY($1)
			GROUP BY product_id
		) AS totals
		WHERE p.id = totals.product_id`,
		pq.Array(productIDs),
	)
	return err
}

func scanProduct(row scanner) (domainproduct.Product, error) {
	var product domainproduct.Product
	var embedding sql.NullString

	err := row.Scan(
		&product.ID,
		&product.CompanyID,
		&product.BranchID,
		&product.SKU,
		&product.Name,
		&product.Description,
		&product.Category,
		&product.Brand,
		&product.PriceCents,
		&product.Currency,
		&product.Stock,
		&embedding,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		return domainproduct.Product{}, err
	}
	if embedding.Valid {
		product.Embedding, err = parseVector(embedding.String)
		if err != nil {
			return domainproduct.Product{}, err
		}
	}

	return product, nil
}
