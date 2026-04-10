package productports

import (
	"context"

	productdto "github.com/IanStuardo-Dev/backend-crud/internal/application/product/dto"
	domainproduct "github.com/IanStuardo-Dev/backend-crud/internal/domain/product"
)

type ProductCatalogWriter interface {
	Create(ctx context.Context, product *domainproduct.Product) error
	Update(ctx context.Context, product *domainproduct.Product) error
	Delete(ctx context.Context, id int64) error
}

type ProductCatalogReader interface {
	List(ctx context.Context) ([]domainproduct.Product, error)
	GetByID(ctx context.Context, id int64) (*domainproduct.Product, error)
}

type ProductSimilarityReader interface {
	FindNeighbors(ctx context.Context, sourceProductID, companyID int64, limit int, minSimilarity float64) ([]productdto.NeighborOutput, error)
}

type ProductSuggestionFeedbackWriter interface {
	SaveNeighborFeedback(ctx context.Context, input productdto.RecordNeighborFeedbackInput) (productdto.NeighborFeedbackOutput, error)
}

type Embedder interface {
	EmbedText(ctx context.Context, text string) ([]float32, error)
}
