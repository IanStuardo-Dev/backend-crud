package postgressale

import (
	"context"
	"database/sql"
	"errors"
	"sort"

	"github.com/lib/pq"

	saleapp "github.com/IanStuardo-Dev/backend-crud/internal/application/sale"
	domainsale "github.com/IanStuardo-Dev/backend-crud/internal/domain/sale"
)

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) saleapp.Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, sale *domainsale.Sale) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = ensureBranchExists(ctx, tx, sale.CompanyID, sale.BranchID); err != nil {
		return err
	}
	if err = ensureUserExists(ctx, tx, sale.CreatedByUserID); err != nil {
		return err
	}

	productRows, err := loadProductsForSale(ctx, tx, sale.CompanyID, sale.BranchID, sale.Items)
	if err != nil {
		return err
	}

	var totalAmount int64
	for index, item := range sale.Items {
		product, ok := productRows[item.ProductID]
		if !ok {
			return saleapp.ErrInvalidReference
		}
		availableStock := product.StockOnHand - product.ReservedStock
		if availableStock < item.Quantity {
			return saleapp.ErrInsufficientStock
		}

		sale.Items[index].ProductSKU = product.SKU
		sale.Items[index].ProductName = product.Name
		sale.Items[index].UnitPriceCents = product.PriceCents
		sale.Items[index].SubtotalCents = int64(item.Quantity) * product.PriceCents
		totalAmount += sale.Items[index].SubtotalCents
	}

	sale.TotalAmountCents = totalAmount
	err = tx.QueryRowContext(
		ctx,
		`INSERT INTO sales (company_id, branch_id, created_by_user_id, total_amount_cents, created_at)
		VALUES ($1,$2,$3,$4,NOW())
		RETURNING id, created_at`,
		sale.CompanyID,
		sale.BranchID,
		sale.CreatedByUserID,
		sale.TotalAmountCents,
	).Scan(&sale.ID, &sale.CreatedAt)
	if err != nil {
		if isForeignKeyViolation(err) {
			return saleapp.ErrInvalidReference
		}
		return err
	}

	for _, item := range sale.Items {
		if _, err = tx.ExecContext(
			ctx,
			`INSERT INTO sale_items (sale_id, product_id, product_sku, product_name, quantity, unit_price_cents, subtotal_cents)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			sale.ID,
			item.ProductID,
			item.ProductSKU,
			item.ProductName,
			item.Quantity,
			item.UnitPriceCents,
			item.SubtotalCents,
		); err != nil {
			return err
		}

		stockAfter := productRows[item.ProductID].StockOnHand - item.Quantity
		if _, err = tx.ExecContext(
			ctx,
			`UPDATE branch_inventory
			SET stock_on_hand = $1, updated_at = NOW()
			WHERE company_id = $2 AND branch_id = $3 AND product_id = $4`,
			stockAfter,
			sale.CompanyID,
			sale.BranchID,
			item.ProductID,
		); err != nil {
			return err
		}

		if _, err = tx.ExecContext(
			ctx,
			`INSERT INTO inventory_movements (
				company_id,
				branch_id,
				product_id,
				sale_id,
				movement_type,
				quantity_delta,
				stock_after,
				created_by_user_id,
				created_at
			) VALUES ($1,$2,$3,$4,'sale',$5,$6,$7,NOW())`,
			sale.CompanyID,
			sale.BranchID,
			item.ProductID,
			sale.ID,
			-item.Quantity,
			stockAfter,
			sale.CreatedByUserID,
		); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *repository) List(ctx context.Context) ([]domainsale.Sale, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, company_id, branch_id, created_by_user_id, total_amount_cents, created_at
		FROM sales
		ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sales := make([]domainsale.Sale, 0)
	for rows.Next() {
		var sale domainsale.Sale
		if err := rows.Scan(&sale.ID, &sale.CompanyID, &sale.BranchID, &sale.CreatedByUserID, &sale.TotalAmountCents, &sale.CreatedAt); err != nil {
			return nil, err
		}
		sales = append(sales, sale)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return attachSaleItems(ctx, r.db, sales)
}

func (r *repository) GetByID(ctx context.Context, id int64) (*domainsale.Sale, error) {
	var sale domainsale.Sale
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, company_id, branch_id, created_by_user_id, total_amount_cents, created_at
		FROM sales
		WHERE id = $1`,
		id,
	).Scan(&sale.ID, &sale.CompanyID, &sale.BranchID, &sale.CreatedByUserID, &sale.TotalAmountCents, &sale.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	sales, err := attachSaleItems(ctx, r.db, []domainsale.Sale{sale})
	if err != nil {
		return nil, err
	}
	return &sales[0], nil
}

type saleProductRow struct {
	ID            int64
	SKU           string
	Name          string
	PriceCents    int64
	StockOnHand   int
	ReservedStock int
}

func loadProductsForSale(ctx context.Context, tx *sql.Tx, companyID, branchID int64, items []domainsale.Item) (map[int64]saleProductRow, error) {
	productIDs := make([]int64, 0, len(items))
	for _, item := range items {
		productIDs = append(productIDs, item.ProductID)
	}
	sort.Slice(productIDs, func(i, j int) bool { return productIDs[i] < productIDs[j] })

	rows, err := tx.QueryContext(
		ctx,
		`SELECT p.id, p.sku, p.name, p.price_cents, bi.stock_on_hand, bi.reserved_stock
		FROM branch_inventory bi
		INNER JOIN products p ON p.id = bi.product_id AND p.company_id = bi.company_id
		WHERE bi.company_id = $1 AND bi.branch_id = $2 AND bi.product_id = ANY($3)
		FOR UPDATE`,
		companyID,
		branchID,
		pq.Array(productIDs),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make(map[int64]saleProductRow, len(items))
	for rows.Next() {
		var product saleProductRow
		if err := rows.Scan(&product.ID, &product.SKU, &product.Name, &product.PriceCents, &product.StockOnHand, &product.ReservedStock); err != nil {
			return nil, err
		}
		products[product.ID] = product
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func ensureBranchExists(ctx context.Context, tx *sql.Tx, companyID, branchID int64) error {
	var exists bool
	err := tx.QueryRowContext(
		ctx,
		`SELECT EXISTS (
			SELECT 1
			FROM branches
			WHERE company_id = $1 AND id = $2
		)`,
		companyID,
		branchID,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return saleapp.ErrInvalidReference
	}
	return nil
}

func ensureUserExists(ctx context.Context, tx *sql.Tx, userID int64) error {
	var exists bool
	err := tx.QueryRowContext(
		ctx,
		`SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE id = $1
		)`,
		userID,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return saleapp.ErrInvalidReference
	}
	return nil
}

type queryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func attachSaleItems(ctx context.Context, q queryer, sales []domainsale.Sale) ([]domainsale.Sale, error) {
	if len(sales) == 0 {
		return sales, nil
	}

	saleIndex := make(map[int64]int, len(sales))
	saleIDs := make([]int64, 0, len(sales))
	for index, sale := range sales {
		saleIndex[sale.ID] = index
		saleIDs = append(saleIDs, sale.ID)
	}

	rows, err := q.QueryContext(
		ctx,
		`SELECT sale_id, product_id, product_sku, product_name, quantity, unit_price_cents, subtotal_cents
		FROM sale_items
		WHERE sale_id = ANY($1)
		ORDER BY sale_id, id`,
		pq.Array(saleIDs),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			saleID int64
			item   domainsale.Item
		)
		if err := rows.Scan(&saleID, &item.ProductID, &item.ProductSKU, &item.ProductName, &item.Quantity, &item.UnitPriceCents, &item.SubtotalCents); err != nil {
			return nil, err
		}
		index := saleIndex[saleID]
		sales[index].Items = append(sales[index].Items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sales, nil
}

func isForeignKeyViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23503"
}
