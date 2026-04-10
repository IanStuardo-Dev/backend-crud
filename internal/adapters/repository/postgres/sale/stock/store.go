package salestockpg

import (
	"context"
	"database/sql"
	"sort"

	postgresshared "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/shared"
	saleports "github.com/IanStuardo-Dev/backend-crud/internal/application/sale/ports"
	domainsale "github.com/IanStuardo-Dev/backend-crud/internal/domain/sale"
	"github.com/lib/pq"
)

type Store struct {
	DB *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{DB: db}
}

func (s *Store) LoadForSale(ctx context.Context, companyID, branchID int64, items []domainsale.Item, lock bool) (map[int64]saleports.StockSnapshot, error) {
	productIDs := make([]int64, 0, len(items))
	for _, item := range items {
		productIDs = append(productIDs, item.ProductID)
	}
	sort.Slice(productIDs, func(i, j int) bool { return productIDs[i] < productIDs[j] })

	query := `SELECT p.id, p.sku, p.name, p.price_cents, bi.stock_on_hand, bi.reserved_stock
		FROM branch_inventory bi
		INNER JOIN products p ON p.id = bi.product_id AND p.company_id = bi.company_id
		WHERE bi.company_id = $1 AND bi.branch_id = $2 AND bi.product_id = ANY($3)`
	if lock {
		query += "\n\t\tFOR UPDATE"
	}

	rows, err := postgresshared.Queryer(ctx, s.DB).QueryContext(ctx, query, companyID, branchID, pq.Array(productIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make(map[int64]saleports.StockSnapshot, len(items))
	for rows.Next() {
		var product saleports.StockSnapshot
		if err := rows.Scan(&product.ProductID, &product.SKU, &product.Name, &product.PriceCents, &product.StockOnHand, &product.ReservedStock); err != nil {
			return nil, err
		}
		product.AvailableStock = product.StockOnHand - product.ReservedStock
		products[product.ProductID] = product
	}

	return products, rows.Err()
}

func (s *Store) SetStockOnHand(ctx context.Context, companyID, branchID, productID int64, stockOnHand int) error {
	_, err := postgresshared.Queryer(ctx, s.DB).ExecContext(
		ctx,
		`UPDATE branch_inventory
		SET stock_on_hand = $1, updated_at = NOW()
		WHERE company_id = $2 AND branch_id = $3 AND product_id = $4`,
		stockOnHand,
		companyID,
		branchID,
		productID,
	)
	return err
}
