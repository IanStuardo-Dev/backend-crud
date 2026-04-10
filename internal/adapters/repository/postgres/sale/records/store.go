package salerecordspg

import (
	"context"
	"database/sql"
	"errors"

	postgresshared "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/shared"
	saleapp "github.com/IanStuardo-Dev/backend-crud/internal/application/sale"
	domainsale "github.com/IanStuardo-Dev/backend-crud/internal/domain/sale"
	"github.com/lib/pq"
)

type Store struct {
	DB *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{DB: db}
}

func (s *Store) Create(ctx context.Context, sale *domainsale.Sale) error {
	exec := postgresshared.Queryer(ctx, s.DB)

	err := exec.QueryRowContext(
		ctx,
		`INSERT INTO sales (company_id, branch_id, created_by_user_id, total_amount_cents, created_at)
		VALUES ($1,$2,$3,$4,NOW())
		RETURNING id, created_at`,
		sale.CompanyID,
		sale.BranchID,
		sale.CreatedByUserID,
		sale.TotalAmountCents,
	).Scan(&sale.ID, &sale.CreatedAt)
	if postgresshared.IsForeignKeyViolation(err) {
		return saleapp.ErrInvalidReference
	}
	if err != nil {
		return err
	}

	for _, item := range sale.Items {
		if _, err := exec.ExecContext(
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
	}

	return nil
}

func (s *Store) List(ctx context.Context) ([]domainsale.Sale, error) {
	rows, err := s.DB.QueryContext(
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

	return attachSaleItems(ctx, s.DB, sales)
}

func (s *Store) GetByID(ctx context.Context, id int64) (*domainsale.Sale, error) {
	var sale domainsale.Sale
	err := s.DB.QueryRowContext(
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

	sales, err := attachSaleItems(ctx, s.DB, []domainsale.Sale{sale})
	if err != nil {
		return nil, err
	}
	return &sales[0], nil
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

	return sales, rows.Err()
}
