package productcatalogpg

import (
	"context"
	"database/sql"
	"errors"

	domainproduct "github.com/IanStuardo-Dev/backend-crud/internal/domain/product"
)

func (s *Store) List(ctx context.Context) ([]domainproduct.Product, error) {
	rows, err := s.DB.QueryContext(
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

	return products, rows.Err()
}

func (s *Store) GetByID(ctx context.Context, id int64) (*domainproduct.Product, error) {
	row := s.DB.QueryRowContext(
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
