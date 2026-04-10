package productcatalogpg

import (
	"database/sql"

	domainproduct "github.com/IanStuardo-Dev/backend-crud/internal/domain/product"
)

type scanner interface {
	Scan(dest ...any) error
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
