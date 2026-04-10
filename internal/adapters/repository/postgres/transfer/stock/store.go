package transferstockpg

import (
	"context"
	"database/sql"
	"sort"

	postgresshared "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/shared"
	transferports "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer/ports"
	domaintransfer "github.com/IanStuardo-Dev/backend-crud/internal/domain/transfer"
	"github.com/lib/pq"
)

type Store struct {
	DB *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{DB: db}
}

func (s *Store) LoadForTransfer(ctx context.Context, companyID, branchID int64, items []domaintransfer.Item, lock bool) (map[int64]transferports.StockSnapshot, error) {
	productIDs := make([]int64, 0, len(items))
	for _, item := range items {
		productIDs = append(productIDs, item.ProductID)
	}
	sort.Slice(productIDs, func(i, j int) bool { return productIDs[i] < productIDs[j] })

	query := `SELECT bi.product_id, p.sku, p.name, bi.stock_on_hand, bi.reserved_stock, (bi.stock_on_hand - bi.reserved_stock) AS available_stock
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

	inventoryRows := make(map[int64]transferports.StockSnapshot, len(items))
	for rows.Next() {
		var row transferports.StockSnapshot
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
			inventoryRows[item.ProductID] = transferports.StockSnapshot{ProductID: item.ProductID}
		}
	}

	return inventoryRows, nil
}

func (s *Store) PutStockOnHand(ctx context.Context, companyID, branchID, productID int64, stockOnHand int) error {
	_, err := postgresshared.Queryer(ctx, s.DB).ExecContext(
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
		companyID,
		branchID,
		productID,
		stockOnHand,
	)
	return err
}
