package productapp

import (
	"context"

	domainproduct "github.com/IanStuardo-Dev/backend-crud/internal/domain/product"
)

// Repository defines the persistence contract required by the use case.
type Repository interface {
	Create(ctx context.Context, product *domainproduct.Product) error
	List(ctx context.Context) ([]domainproduct.Product, error)
	GetByID(ctx context.Context, id int64) (*domainproduct.Product, error)
	FindNeighbors(ctx context.Context, sourceProductID, companyID int64, limit int, minSimilarity float64) ([]NeighborOutput, error)
	SaveNeighborFeedback(ctx context.Context, input RecordNeighborFeedbackInput) (NeighborFeedbackOutput, error)
	Update(ctx context.Context, product *domainproduct.Product) error
	Delete(ctx context.Context, id int64) error
}

type Embedder interface {
	EmbedText(ctx context.Context, text string) ([]float32, error)
}

// UseCase defines the application operations for products.
type UseCase interface {
	Create(ctx context.Context, input CreateInput) (Output, error)
	List(ctx context.Context) ([]Output, error)
	GetByID(ctx context.Context, id int64) (Output, error)
	FindNeighbors(ctx context.Context, input FindNeighborsInput) (FindNeighborsOutput, error)
	RecordNeighborFeedback(ctx context.Context, input RecordNeighborFeedbackInput) (NeighborFeedbackOutput, error)
	Update(ctx context.Context, input UpdateInput) (Output, error)
	Delete(ctx context.Context, id int64) error
}
