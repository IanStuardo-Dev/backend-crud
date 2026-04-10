package productcatalogpg

import (
	"context"
	"database/sql"
	"errors"

	postgresshared "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/shared"
	productapp "github.com/IanStuardo-Dev/backend-crud/internal/application/product"
	domainproduct "github.com/IanStuardo-Dev/backend-crud/internal/domain/product"
)

func (s *Store) Update(ctx context.Context, product *domainproduct.Product) (err error) {
	tx, err := s.DB.BeginTx(ctx, nil)
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
	if postgresshared.IsUniqueViolation(err) {
		return productapp.ErrConflict
	}
	if postgresshared.IsForeignKeyViolation(err) {
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
		if postgresshared.IsUniqueViolation(err) {
			return productapp.ErrConflict
		}
		if postgresshared.IsForeignKeyViolation(err) {
			return productapp.ErrInvalidReference
		}
		return err
	}

	if err = syncProductTotalStock(ctx, tx, []int64{product.ID}); err != nil {
		return err
	}

	return tx.Commit()
}
