package postgresproduct

import (
	"database/sql"

	productcatalogpg "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/product/catalog"
	productfeedbackpg "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/product/feedback"
	productsimilaritypg "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/product/similarity"
)

func NewCatalogStore(db *sql.DB) *productcatalogpg.Store {
	return productcatalogpg.NewStore(db)
}

func NewSimilarityStore(db *sql.DB) *productsimilaritypg.Store {
	return productsimilaritypg.NewStore(db)
}

func NewFeedbackStore(db *sql.DB) *productfeedbackpg.Store {
	return productfeedbackpg.NewStore(db)
}
