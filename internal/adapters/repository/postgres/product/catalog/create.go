package productcatalogpg

import (
	"context"

	postgresshared "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/shared"
	productapp "github.com/IanStuardo-Dev/backend-crud/internal/application/product"
	domainproduct "github.com/IanStuardo-Dev/backend-crud/internal/domain/product"
)

func (s *Store) Create(ctx context.Context, product *domainproduct.Product) (err error) {
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
		) VALUES ($1,$2,$3,$4,0,NOW(),NOW())`,
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

	return tx.Commit()
}
