package productapp

import (
	"context"
	"errors"
	"testing"
	"time"

	domainproduct "github.com/IanStuardo-Dev/backend-crud/internal/domain/product"
)

type stubRepository struct {
	createFn               func(context.Context, *domainproduct.Product) error
	listFn                 func(context.Context) ([]domainproduct.Product, error)
	getByIDFn              func(context.Context, int64) (*domainproduct.Product, error)
	findNeighborsFn        func(context.Context, int64, int64, int, float64) ([]NeighborOutput, error)
	saveNeighborFeedbackFn func(context.Context, RecordNeighborFeedbackInput) (NeighborFeedbackOutput, error)
	updateFn               func(context.Context, *domainproduct.Product) error
	deleteFn               func(context.Context, int64) error
}

func (s stubRepository) Create(ctx context.Context, product *domainproduct.Product) error {
	if s.createFn != nil {
		return s.createFn(ctx, product)
	}

	return nil
}

func (s stubRepository) List(ctx context.Context) ([]domainproduct.Product, error) {
	if s.listFn != nil {
		return s.listFn(ctx)
	}

	return nil, nil
}

func (s stubRepository) GetByID(ctx context.Context, id int64) (*domainproduct.Product, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(ctx, id)
	}

	return nil, nil
}

func (s stubRepository) FindNeighbors(ctx context.Context, sourceProductID, companyID int64, limit int, minSimilarity float64) ([]NeighborOutput, error) {
	if s.findNeighborsFn != nil {
		return s.findNeighborsFn(ctx, sourceProductID, companyID, limit, minSimilarity)
	}

	return nil, nil
}

func (s stubRepository) SaveNeighborFeedback(ctx context.Context, input RecordNeighborFeedbackInput) (NeighborFeedbackOutput, error) {
	if s.saveNeighborFeedbackFn != nil {
		return s.saveNeighborFeedbackFn(ctx, input)
	}

	return NeighborFeedbackOutput{}, nil
}

func (s stubRepository) Update(ctx context.Context, product *domainproduct.Product) error {
	if s.updateFn != nil {
		return s.updateFn(ctx, product)
	}

	return nil
}

func (s stubRepository) Delete(ctx context.Context, id int64) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, id)
	}

	return nil
}

type stubEmbedder struct {
	embedTextFn func(context.Context, string) ([]float32, error)
}

func (s stubEmbedder) EmbedText(ctx context.Context, text string) ([]float32, error) {
	if s.embedTextFn != nil {
		return s.embedTextFn(ctx, text)
	}

	return makeEmbedding(), nil
}

func TestUseCaseCreateNormalizesAndValidates(t *testing.T) {
	now := time.Now().UTC()
	var createdProduct domainproduct.Product

	useCase := NewUseCase(stubRepository{
		createFn: func(_ context.Context, product *domainproduct.Product) error {
			createdProduct = *product
			product.ID = 9
			product.CreatedAt = now
			product.UpdatedAt = now
			return nil
		},
	}, stubEmbedder{})

	output, err := useCase.Create(context.Background(), CreateInput{
		CompanyID:  1,
		BranchID:   1,
		SKU:        " sku-001 ",
		Name:       "  Noise Cancelling Headphones  ",
		Category:   "  audio ",
		Brand:      "  Acme ",
		PriceCents: 15999,
		Currency:   " usd ",
		Stock:      12,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if createdProduct.SKU != "SKU-001" {
		t.Fatalf("expected normalized SKU, got %q", createdProduct.SKU)
	}
	if createdProduct.Currency != "USD" {
		t.Fatalf("expected normalized currency, got %q", createdProduct.Currency)
	}
	if output.ID != 9 {
		t.Fatalf("expected generated ID to be returned, got %d", output.ID)
	}
	if output.CreatedAt != now {
		t.Fatalf("expected created_at to be propagated")
	}
}

func TestUseCaseCreateRejectsInvalidEmbedding(t *testing.T) {
	called := false

	useCase := NewUseCase(stubRepository{
		createFn: func(_ context.Context, product *domainproduct.Product) error {
			called = true
			return nil
		},
	}, nil)

	_, err := useCase.Create(context.Background(), CreateInput{
		CompanyID: 1,
		BranchID:  1,
		SKU:       "SKU-001",
		Name:      "Headphones",
		Category:  "audio",
		Currency:  "USD",
		Embedding: []float32{0.1, 0.2},
	})

	var validationErr domainproduct.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if validationErr.Field != "embedding" {
		t.Fatalf("expected embedding validation, got %q", validationErr.Field)
	}
	if called {
		t.Fatal("expected repository Create not to be called on invalid input")
	}
}

func TestUseCaseGetByIDReturnsNotFound(t *testing.T) {
	useCase := NewUseCase(stubRepository{
		getByIDFn: func(context.Context, int64) (*domainproduct.Product, error) {
			return nil, nil
		},
	}, nil)

	_, err := useCase.GetByID(context.Background(), 10)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUseCaseRecordNeighborFeedbackNormalizesAndStores(t *testing.T) {
	now := time.Now().UTC()
	products := map[int64]domainproduct.Product{
		11: {
			ID:        11,
			CompanyID: 3,
			BranchID:  9,
			SKU:       "SKU-011",
			Name:      "Cafe Clasico",
			Category:  "abarrotes",
			Currency:  "CLP",
			Embedding: makeEmbedding(),
		},
		18: {
			ID:        18,
			CompanyID: 3,
			BranchID:  9,
			SKU:       "SKU-018",
			Name:      "Cafe Premium",
			Category:  "abarrotes",
			Currency:  "CLP",
			Embedding: makeEmbedding(),
		},
	}

	var savedInput RecordNeighborFeedbackInput
	useCase := NewUseCase(stubRepository{
		getByIDFn: func(_ context.Context, id int64) (*domainproduct.Product, error) {
			product, ok := products[id]
			if !ok {
				return nil, nil
			}
			productCopy := product
			return &productCopy, nil
		},
		saveNeighborFeedbackFn: func(_ context.Context, input RecordNeighborFeedbackInput) (NeighborFeedbackOutput, error) {
			savedInput = input
			return NeighborFeedbackOutput{
				SourceProductID:    input.SourceProductID,
				SuggestedProductID: input.SuggestedProductID,
				CompanyID:          input.CompanyID,
				BranchID:           input.BranchID,
				UserID:             input.UserID,
				Action:             input.Action,
				Note:               input.Note,
				CreatedAt:          now,
				UpdatedAt:          now,
			}, nil
		},
	}, nil)

	output, err := useCase.RecordNeighborFeedback(context.Background(), RecordNeighborFeedbackInput{
		SourceProductID:    11,
		SuggestedProductID: 18,
		BranchID:           9,
		UserID:             77,
		Action:             " Accepted ",
		Note:               "  mejor reemplazo para caja rapida  ",
	})
	if err != nil {
		t.Fatalf("RecordNeighborFeedback() error = %v", err)
	}

	if savedInput.CompanyID != 3 {
		t.Fatalf("expected company_id 3, got %d", savedInput.CompanyID)
	}
	if savedInput.Action != "accepted" {
		t.Fatalf("expected normalized action, got %q", savedInput.Action)
	}
	if savedInput.Note != "mejor reemplazo para caja rapida" {
		t.Fatalf("expected trimmed note, got %q", savedInput.Note)
	}
	if output.CompanyID != 3 || output.CreatedAt != now || output.UpdatedAt != now {
		t.Fatalf("unexpected output %#v", output)
	}
}

func TestUseCaseRecordNeighborFeedbackRejectsProductsFromDifferentCompanies(t *testing.T) {
	useCase := NewUseCase(stubRepository{
		getByIDFn: func(_ context.Context, id int64) (*domainproduct.Product, error) {
			switch id {
			case 11:
				return &domainproduct.Product{ID: 11, CompanyID: 1, BranchID: 1, SKU: "SKU-011", Name: "Producto A", Category: "office", Currency: "USD"}, nil
			case 18:
				return &domainproduct.Product{ID: 18, CompanyID: 2, BranchID: 2, SKU: "SKU-018", Name: "Producto B", Category: "office", Currency: "USD"}, nil
			default:
				return nil, nil
			}
		},
	}, nil)

	_, err := useCase.RecordNeighborFeedback(context.Background(), RecordNeighborFeedbackInput{
		SourceProductID:    11,
		SuggestedProductID: 18,
		BranchID:           1,
		UserID:             7,
		Action:             "accepted",
	})
	if !errors.Is(err, ErrInvalidReference) {
		t.Fatalf("expected ErrInvalidReference, got %v", err)
	}
}

func TestUseCaseRecordNeighborFeedbackRejectsInvalidAction(t *testing.T) {
	called := false
	useCase := NewUseCase(stubRepository{
		getByIDFn: func(_ context.Context, id int64) (*domainproduct.Product, error) {
			return &domainproduct.Product{ID: id, CompanyID: 1, BranchID: 1, SKU: "SKU-001", Name: "Producto", Category: "office", Currency: "USD"}, nil
		},
		saveNeighborFeedbackFn: func(context.Context, RecordNeighborFeedbackInput) (NeighborFeedbackOutput, error) {
			called = true
			return NeighborFeedbackOutput{}, nil
		},
	}, nil)

	_, err := useCase.RecordNeighborFeedback(context.Background(), RecordNeighborFeedbackInput{
		SourceProductID:    1,
		SuggestedProductID: 2,
		BranchID:           1,
		UserID:             7,
		Action:             "maybe",
	})

	var validationErr domainproduct.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if validationErr.Field != "action" {
		t.Fatalf("expected action validation, got %q", validationErr.Field)
	}
	if called {
		t.Fatal("expected repository SaveNeighborFeedback not to be called on invalid action")
	}
}

func TestUseCaseUpdateNormalizesAndValidates(t *testing.T) {
	now := time.Now().UTC()
	var updatedProduct domainproduct.Product

	useCase := NewUseCase(stubRepository{
		getByIDFn: func(context.Context, int64) (*domainproduct.Product, error) {
			return &domainproduct.Product{
				ID:          3,
				CompanyID:   1,
				BranchID:    1,
				SKU:         "SKU-003",
				Name:        "Mechanical Keyboard",
				Description: "tactile switches",
				Category:    "peripherals",
				Brand:       "ACME",
				PriceCents:  8999,
				Currency:    "USD",
				Stock:       9,
				Embedding:   makeEmbedding(),
			}, nil
		},
		updateFn: func(_ context.Context, product *domainproduct.Product) error {
			updatedProduct = *product
			product.CreatedAt = now.Add(-time.Hour)
			product.UpdatedAt = now
			return nil
		},
	}, stubEmbedder{})

	output, err := useCase.Update(context.Background(), UpdateInput{
		ID:          3,
		CompanyID:   1,
		BranchID:    1,
		SKU:         " sku-003 ",
		Name:        "  Mechanical Keyboard ",
		Description: " tactile switches ",
		Category:    " peripherals ",
		Brand:       " acme ",
		PriceCents:  8999,
		Currency:    " usd ",
		Stock:       7,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if updatedProduct.SKU != "SKU-003" {
		t.Fatalf("expected normalized sku, got %q", updatedProduct.SKU)
	}
	if updatedProduct.Name != "Mechanical Keyboard" {
		t.Fatalf("expected normalized name, got %q", updatedProduct.Name)
	}
	if output.UpdatedAt != now {
		t.Fatalf("expected updated_at to be propagated")
	}
}

func TestUseCaseDeleteRejectsInvalidID(t *testing.T) {
	called := false

	useCase := NewUseCase(stubRepository{
		deleteFn: func(_ context.Context, id int64) error {
			called = true
			return nil
		},
	}, nil)

	err := useCase.Delete(context.Background(), 0)

	var validationErr domainproduct.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if validationErr.Field != "id" {
		t.Fatalf("expected id field validation, got %q", validationErr.Field)
	}
	if called {
		t.Fatal("expected repository Delete not to be called on invalid id")
	}
}

func TestUseCaseDeletePropagatesNotFound(t *testing.T) {
	useCase := NewUseCase(stubRepository{
		deleteFn: func(_ context.Context, id int64) error {
			return ErrNotFound
		},
	}, nil)

	err := useCase.Delete(context.Background(), 10)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUseCaseReadOperationsDelegate(t *testing.T) {
	now := time.Now().UTC()
	expectedProducts := []domainproduct.Product{{
		ID:         1,
		CompanyID:  1,
		BranchID:   1,
		SKU:        "SKU-001",
		Name:       "Headphones",
		Category:   "audio",
		PriceCents: 15999,
		Currency:   "USD",
		Stock:      4,
		CreatedAt:  now,
		UpdatedAt:  now,
	}}
	expectedProduct := &domainproduct.Product{
		ID:         2,
		CompanyID:  1,
		BranchID:   1,
		SKU:        "SKU-002",
		Name:       "Mouse",
		Category:   "peripherals",
		PriceCents: 4999,
		Currency:   "USD",
		Stock:      8,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	useCase := NewUseCase(stubRepository{
		listFn: func(context.Context) ([]domainproduct.Product, error) {
			return expectedProducts, nil
		},
		getByIDFn: func(_ context.Context, id int64) (*domainproduct.Product, error) {
			if id != 2 {
				t.Fatalf("expected id 2, got %d", id)
			}
			return expectedProduct, nil
		},
	}, nil)

	products, err := useCase.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(products) != 1 || products[0].SKU != expectedProducts[0].SKU {
		t.Fatalf("unexpected products: %#v", products)
	}

	product, err := useCase.GetByID(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if product.Name != expectedProduct.Name {
		t.Fatalf("unexpected product: %#v", product)
	}
}

func TestUseCaseCreateGeneratesEmbeddingWhenMissing(t *testing.T) {
	var generatedText string

	useCase := NewUseCase(stubRepository{
		createFn: func(_ context.Context, product *domainproduct.Product) error {
			if len(product.Embedding) != domainproduct.EmbeddingDimensions {
				t.Fatalf("expected generated embedding, got len=%d", len(product.Embedding))
			}
			return nil
		},
	}, stubEmbedder{
		embedTextFn: func(_ context.Context, text string) ([]float32, error) {
			generatedText = text
			return makeEmbedding(), nil
		},
	})

	_, err := useCase.Create(context.Background(), CreateInput{
		CompanyID:   1,
		BranchID:    1,
		SKU:         "sku-001",
		Name:        "Headphones",
		Description: "Noise cancelling",
		Category:    "audio",
		Brand:       "Acme",
		PriceCents:  15999,
		Currency:    "USD",
		Stock:       2,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if generatedText == "" {
		t.Fatal("expected embedding text to be generated")
	}
}

func TestUseCaseCreateFailsWhenEmbeddingProviderIsMissing(t *testing.T) {
	useCase := NewUseCase(stubRepository{}, nil)

	_, err := useCase.Create(context.Background(), CreateInput{
		CompanyID:  1,
		BranchID:   1,
		SKU:        "SKU-001",
		Name:       "Headphones",
		Category:   "audio",
		Currency:   "USD",
		PriceCents: 1000,
		Stock:      1,
	})
	if !errors.Is(err, ErrEmbeddingUnavailable) {
		t.Fatalf("expected ErrEmbeddingUnavailable, got %v", err)
	}
}

func TestUseCaseFindNeighborsReturnsRankedNeighbors(t *testing.T) {
	useCase := NewUseCase(stubRepository{
		getByIDFn: func(context.Context, int64) (*domainproduct.Product, error) {
			return &domainproduct.Product{
				ID:        4,
				CompanyID: 1,
				Name:      "Wireless Mouse",
				Embedding: makeEmbedding(),
			}, nil
		},
		findNeighborsFn: func(_ context.Context, sourceProductID, companyID int64, limit int, minSimilarity float64) ([]NeighborOutput, error) {
			if sourceProductID != 4 || companyID != 1 {
				t.Fatalf("unexpected query scope: product=%d company=%d", sourceProductID, companyID)
			}
			if limit != 5 {
				t.Fatalf("expected default limit 5, got %d", limit)
			}
			if minSimilarity != 0.35 {
				t.Fatalf("expected minSimilarity 0.35, got %v", minSimilarity)
			}

			return []NeighborOutput{
				{
					ProductID:   13,
					SKU:         "SEED-PRD-010",
					Name:        "Wireless Ergonomic Mouse",
					Distance:    0.3028764,
					PriceCents:  25990,
					Currency:    "CLP",
					Description: "Mouse inalambrico ergonomico.",
					Category:    "peripherals",
					Brand:       "Acme",
				},
				{
					ProductID:   8,
					SKU:         "SEED-PRD-005",
					Name:        "Laptop Stand Aluminum",
					Distance:    0.6437432,
					PriceCents:  24990,
					Currency:    "CLP",
					Description: "Soporte de aluminio.",
					Category:    "office",
					Brand:       "Acme",
				},
			}, nil
		},
	}, nil)

	output, err := useCase.FindNeighbors(context.Background(), FindNeighborsInput{
		ProductID:     4,
		MinSimilarity: 0.35,
	})
	if err != nil {
		t.Fatalf("FindNeighbors() error = %v", err)
	}
	if output.SourceProductID != 4 || output.SourceProductName != "Wireless Mouse" {
		t.Fatalf("unexpected source metadata: %#v", output)
	}
	if len(output.Neighbors) != 2 {
		t.Fatalf("expected 2 neighbors, got %d", len(output.Neighbors))
	}
	if output.Neighbors[0].SimilarityPercentage != 69.71 {
		t.Fatalf("expected rounded similarity 69.71, got %v", output.Neighbors[0].SimilarityPercentage)
	}
	if output.Neighbors[0].Distance != 0.302876 {
		t.Fatalf("expected rounded distance 0.302876, got %v", output.Neighbors[0].Distance)
	}
}

func TestUseCaseFindNeighborsRejectsMissingSourceEmbedding(t *testing.T) {
	useCase := NewUseCase(stubRepository{
		getByIDFn: func(context.Context, int64) (*domainproduct.Product, error) {
			return &domainproduct.Product{
				ID:        4,
				CompanyID: 1,
				Name:      "Wireless Mouse",
			}, nil
		},
	}, nil)

	_, err := useCase.FindNeighbors(context.Background(), FindNeighborsInput{ProductID: 4})
	if !errors.Is(err, ErrSourceEmbeddingMissing) {
		t.Fatalf("expected ErrSourceEmbeddingMissing, got %v", err)
	}
}

func TestUseCaseFindNeighborsRejectsInvalidMinSimilarity(t *testing.T) {
	useCase := NewUseCase(stubRepository{}, nil)

	_, err := useCase.FindNeighbors(context.Background(), FindNeighborsInput{
		ProductID:     4,
		MinSimilarity: 1.5,
	})

	var validationErr domainproduct.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if validationErr.Field != "min_similarity" {
		t.Fatalf("expected min_similarity validation, got %q", validationErr.Field)
	}
}

func TestUseCaseUpdateReusesExistingEmbeddingWhenTextDoesNotChange(t *testing.T) {
	existingEmbedding := makeEmbedding()
	var updatedProduct domainproduct.Product

	useCase := NewUseCase(stubRepository{
		getByIDFn: func(context.Context, int64) (*domainproduct.Product, error) {
			return &domainproduct.Product{
				ID:          5,
				CompanyID:   1,
				BranchID:    1,
				SKU:         "SKU-005",
				Name:        "Gaming Mouse",
				Description: "Wireless mouse",
				Category:    "peripherals",
				Brand:       "Acme",
				PriceCents:  5000,
				Currency:    "USD",
				Stock:       4,
				Embedding:   existingEmbedding,
			}, nil
		},
		updateFn: func(_ context.Context, product *domainproduct.Product) error {
			updatedProduct = *product
			return nil
		},
	}, nil)

	_, err := useCase.Update(context.Background(), UpdateInput{
		ID:          5,
		CompanyID:   1,
		BranchID:    1,
		SKU:         "SKU-005",
		Name:        "Gaming Mouse",
		Description: "Wireless mouse",
		Category:    "peripherals",
		Brand:       "Acme",
		PriceCents:  5000,
		Currency:    "USD",
		Stock:       8,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if len(updatedProduct.Embedding) != len(existingEmbedding) {
		t.Fatalf("expected existing embedding to be reused, got len=%d", len(updatedProduct.Embedding))
	}
}

func makeEmbedding() []float32 {
	embedding := make([]float32, domainproduct.EmbeddingDimensions)
	for index := range embedding {
		embedding[index] = 0.001
	}

	return embedding
}
