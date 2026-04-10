package productcatalogpg

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

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
