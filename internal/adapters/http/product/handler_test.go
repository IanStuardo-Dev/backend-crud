package producthttp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	producthttp "github.com/example/crud/internal/adapters/http/product"
	productapp "github.com/example/crud/internal/application/product"
	domainproduct "github.com/example/crud/internal/domain/product"
	"github.com/example/crud/internal/infrastructure/http/router"
)

type memoryProductRepository struct {
	nextID   int64
	products map[int64]domainproduct.Product
}

func newMemoryProductRepository() *memoryProductRepository {
	return &memoryProductRepository{
		nextID:   1,
		products: make(map[int64]domainproduct.Product),
	}
}

func (r *memoryProductRepository) Create(_ context.Context, product *domainproduct.Product) error {
	for _, existing := range r.products {
		if existing.BranchID == product.BranchID && existing.SKU == product.SKU {
			return productapp.ErrConflict
		}
	}

	now := time.Now().UTC().Round(0)
	product.ID = r.nextID
	product.CreatedAt = now
	product.UpdatedAt = now
	r.nextID++
	r.products[product.ID] = *product

	return nil
}

func (r *memoryProductRepository) List(context.Context) ([]domainproduct.Product, error) {
	products := make([]domainproduct.Product, 0, len(r.products))
	for id := int64(1); id < r.nextID; id++ {
		if product, ok := r.products[id]; ok {
			products = append(products, product)
		}
	}

	return products, nil
}

func (r *memoryProductRepository) GetByID(_ context.Context, id int64) (*domainproduct.Product, error) {
	product, ok := r.products[id]
	if !ok {
		return nil, nil
	}

	productCopy := product
	return &productCopy, nil
}

func (r *memoryProductRepository) FindNeighbors(_ context.Context, sourceProductID, companyID int64, limit int, minSimilarity float64) ([]productapp.NeighborOutput, error) {
	source, ok := r.products[sourceProductID]
	if !ok {
		return nil, nil
	}

	neighbors := make([]productapp.NeighborOutput, 0)
	for _, product := range r.products {
		if product.ID == sourceProductID || product.CompanyID != companyID || len(product.Embedding) == 0 {
			continue
		}

		distance := fakeDistance(source.Name, product.Name)
		similarity := 1 - distance
		if similarity < minSimilarity {
			continue
		}

		neighbors = append(neighbors, productapp.NeighborOutput{
			ProductID:   product.ID,
			SKU:         product.SKU,
			Name:        product.Name,
			Description: product.Description,
			Category:    product.Category,
			Brand:       product.Brand,
			PriceCents:  product.PriceCents,
			Currency:    product.Currency,
			Distance:    distance,
		})
	}

	sort.Slice(neighbors, func(i, j int) bool {
		return neighbors[i].Distance < neighbors[j].Distance
	})
	if limit > len(neighbors) {
		limit = len(neighbors)
	}

	return neighbors[:limit], nil
}

func (r *memoryProductRepository) Update(_ context.Context, product *domainproduct.Product) error {
	existing, ok := r.products[product.ID]
	if !ok {
		return productapp.ErrNotFound
	}
	for id, saved := range r.products {
		if id != product.ID && saved.BranchID == product.BranchID && saved.SKU == product.SKU {
			return productapp.ErrConflict
		}
	}

	product.CreatedAt = existing.CreatedAt
	product.UpdatedAt = time.Now().UTC().Round(0)
	r.products[product.ID] = *product
	return nil
}

func (r *memoryProductRepository) Delete(_ context.Context, id int64) error {
	if _, ok := r.products[id]; !ok {
		return productapp.ErrNotFound
	}

	delete(r.products, id)
	return nil
}

func TestProductCRUDFlow(t *testing.T) {
	repo := newMemoryProductRepository()
	handler := producthttp.NewHandler(productapp.NewUseCase(repo, testProductEmbedder{}))
	server := router.New(nil, nil, nil, nil, handler, nil, nil)

	createBody := `{"company_id":1,"branch_id":1,"sku":"sku-001","name":"Laptop Stand","description":"Aluminum stand","category":"office","brand":"Acme","price_cents":3499,"currency":"usd","stock":10}`
	createResp := performRequest(t, server, http.MethodPost, "/products", createBody)
	assertStatus(t, createResp, http.StatusCreated)
	assertHeader(t, createResp, "Location", "/products/1")

	var createdBody map[string]any
	decodeBody(t, createResp, &createdBody)
	data := createdBody["data"].(map[string]any)
	if data["sku"] != "SKU-001" {
		t.Fatalf("expected normalized SKU, got %#v", data["sku"])
	}
	if data["currency"] != "USD" {
		t.Fatalf("expected normalized currency, got %#v", data["currency"])
	}
	if data["company_id"] != float64(1) || data["branch_id"] != float64(1) {
		t.Fatalf("expected company_id and branch_id in response, got %#v", data)
	}

	getAllResp := performRequest(t, server, http.MethodGet, "/products", "")
	assertStatus(t, getAllResp, http.StatusOK)

	getByIDResp := performRequest(t, server, http.MethodGet, "/products/1", "")
	assertStatus(t, getByIDResp, http.StatusOK)

	updateBody := `{"company_id":1,"branch_id":1,"sku":"sku-001","name":"Laptop Stand Pro","description":"Aluminum stand","category":"office","brand":"Acme","price_cents":3999,"currency":"USD","stock":8}`
	updateResp := performRequest(t, server, http.MethodPut, "/products/1", updateBody)
	assertStatus(t, updateResp, http.StatusOK)

	deleteResp := performRequest(t, server, http.MethodDelete, "/products/1", "")
	assertStatus(t, deleteResp, http.StatusNoContent)

	notFoundResp := performRequest(t, server, http.MethodGet, "/products/1", "")
	assertStatus(t, notFoundResp, http.StatusNotFound)
	assertProblemResponse(t, notFoundResp, map[string]any{
		"type":   "https://httpstatuses.com/404",
		"title":  "Resource Not Found",
		"status": float64(404),
		"detail": "product not found",
		"path":   "/products/1",
	})
}

func TestProductValidationAndErrors(t *testing.T) {
	repo := newMemoryProductRepository()
	handler := producthttp.NewHandler(productapp.NewUseCase(repo, testProductEmbedder{}))
	server := router.New(nil, nil, nil, nil, handler, nil, nil)

	cases := []struct {
		name       string
		method     string
		path       string
		body       string
		setJSONCT  bool
		statusCode int
		wantBody   map[string]any
		prepare    func(t *testing.T)
	}{
		{
			name:       "reject invalid embedding dimensions",
			method:     http.MethodPost,
			path:       "/products",
			body:       `{"company_id":1,"branch_id":1,"sku":"sku-001","name":"Laptop Stand","category":"office","currency":"USD","embedding":[0.1,0.2]}`,
			setJSONCT:  true,
			statusCode: http.StatusUnprocessableEntity,
			wantBody: map[string]any{
				"type":   "https://httpstatuses.com/422",
				"title":  "Validation Failed",
				"status": float64(422),
				"detail": "request validation failed",
				"path":   "/products",
				"errors": []any{
					map[string]any{
						"field":  "embedding",
						"reason": "embedding must have exactly 1536 dimensions",
					},
				},
			},
		},
		{
			name:       "reject duplicate sku on create",
			method:     http.MethodPost,
			path:       "/products",
			body:       `{"company_id":1,"branch_id":1,"sku":"sku-001","name":"Other Product","category":"office","currency":"USD"}`,
			setJSONCT:  true,
			statusCode: http.StatusConflict,
			wantBody: map[string]any{
				"type":   "https://httpstatuses.com/409",
				"title":  "Conflict",
				"status": float64(409),
				"detail": "a product with the same sku already exists in this branch",
				"path":   "/products",
			},
			prepare: func(t *testing.T) {
				t.Helper()
				err := repo.Create(context.Background(), &domainproduct.Product{
					CompanyID:  1,
					BranchID:   1,
					SKU:        "SKU-001",
					Name:       "Laptop Stand",
					Category:   "office",
					Currency:   "USD",
					PriceCents: 3499,
				})
				if err != nil {
					t.Fatalf("seed product: %v", err)
				}
			},
		},
		{
			name:       "reject missing content type",
			method:     http.MethodPost,
			path:       "/products",
			body:       `{"company_id":1,"branch_id":1,"sku":"sku-001","name":"Laptop Stand","category":"office","currency":"USD"}`,
			statusCode: http.StatusUnsupportedMediaType,
			wantBody: map[string]any{
				"type":   "https://httpstatuses.com/415",
				"title":  "Unsupported Media Type",
				"status": float64(415),
				"detail": "Content-Type must be application/json",
				"path":   "/products",
			},
		},
		{
			name:       "reject invalid identifier",
			method:     http.MethodDelete,
			path:       "/products/abc",
			statusCode: http.StatusNotFound,
			wantBody: map[string]any{
				"type":   "https://httpstatuses.com/404",
				"title":  "Resource Not Found",
				"status": float64(404),
				"detail": "endpoint not found",
				"path":   "/products/abc",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo.nextID = 1
			repo.products = make(map[int64]domainproduct.Product)
			if tc.prepare != nil {
				tc.prepare(t)
			}

			resp := performRequestWithJSONOption(t, server, tc.method, tc.path, tc.body, tc.setJSONCT)
			assertStatus(t, resp, tc.statusCode)
			assertProblemResponse(t, resp, tc.wantBody)
		})
	}
}

func TestProductHandlerMapsUnexpectedErrorsToInternalServerError(t *testing.T) {
	useCase := productapp.NewUseCase(failingRepository{err: errors.New("boom")}, testProductEmbedder{})
	server := router.New(nil, nil, nil, nil, producthttp.NewHandler(useCase), nil, nil)

	resp := performRequest(t, server, http.MethodGet, "/products", "")

	assertStatus(t, resp, http.StatusInternalServerError)
	assertProblemResponse(t, resp, map[string]any{
		"type":   "https://httpstatuses.com/500",
		"title":  "Internal Server Error",
		"status": float64(500),
		"detail": "an unexpected error occurred",
		"path":   "/products",
	})
}

func TestProductCreateReturnsServiceUnavailableWhenEmbeddingProviderIsMissing(t *testing.T) {
	useCase := productapp.NewUseCase(newMemoryProductRepository(), nil)
	server := router.New(nil, nil, nil, nil, producthttp.NewHandler(useCase), nil, nil)

	resp := performRequest(
		t,
		server,
		http.MethodPost,
		"/products",
		`{"company_id":1,"branch_id":1,"sku":"sku-001","name":"Laptop Stand","description":"Aluminum stand","category":"office","brand":"Acme","price_cents":3499,"currency":"USD","stock":10}`,
	)

	assertStatus(t, resp, http.StatusServiceUnavailable)
	assertProblemResponse(t, resp, map[string]any{
		"type":   "https://httpstatuses.com/503",
		"title":  "Service Unavailable",
		"status": float64(503),
		"detail": "product embedding provider is not configured",
		"path":   "/products",
	})
}

func TestProductFindNeighborsFlow(t *testing.T) {
	repo := newMemoryProductRepository()
	handler := producthttp.NewHandler(productapp.NewUseCase(repo, testProductEmbedder{}))
	server := router.New(nil, nil, nil, nil, handler, nil, nil)

	mustCreateProduct(t, server, `{"company_id":1,"branch_id":1,"sku":"seed-prd-001","name":"Wireless Mouse","description":"Mouse inalambrico","category":"peripherals","brand":"Acme","price_cents":19990,"currency":"CLP","stock":10}`)
	mustCreateProduct(t, server, `{"company_id":1,"branch_id":1,"sku":"seed-prd-002","name":"Laptop Cooling Pad","description":"Base de enfriamiento","category":"office","brand":"Northwind","price_cents":27990,"currency":"CLP","stock":10}`)
	mustCreateProduct(t, server, `{"company_id":1,"branch_id":1,"sku":"seed-prd-010","name":"Wireless Ergonomic Mouse","description":"Mouse ergonomico inalambrico","category":"peripherals","brand":"Acme","price_cents":25990,"currency":"CLP","stock":10}`)

	resp := performRequest(t, server, http.MethodGet, "/products/1/neighbors?limit=5&min_similarity=0.20", "")
	assertStatus(t, resp, http.StatusOK)

	var body map[string]any
	decodeBody(t, resp, &body)
	data := body["data"].([]any)
	if len(data) == 0 {
		t.Fatal("expected at least one neighbor")
	}

	first := data[0].(map[string]any)
	if first["product_id"] != float64(3) {
		t.Fatalf("expected product 3 as closest neighbor, got %#v", first["product_id"])
	}
	if first["name"] != "Wireless Ergonomic Mouse" {
		t.Fatalf("unexpected closest neighbor %#v", first["name"])
	}
	if first["similarity_percentage"].(float64) <= 0 {
		t.Fatalf("expected positive similarity, got %#v", first["similarity_percentage"])
	}

	meta := body["meta"].(map[string]any)
	if meta["source_product_id"] != float64(1) || meta["source_product_name"] != "Wireless Mouse" {
		t.Fatalf("unexpected meta %#v", meta)
	}
}

func TestProductFindNeighborsValidationAndErrors(t *testing.T) {
	t.Run("reject invalid query values", func(t *testing.T) {
		repo := newMemoryProductRepository()
		handler := producthttp.NewHandler(productapp.NewUseCase(repo, testProductEmbedder{}))
		server := router.New(nil, nil, nil, nil, handler, nil, nil)

		resp := performRequest(t, server, http.MethodGet, "/products/1/neighbors?limit=nope", "")
		assertStatus(t, resp, http.StatusBadRequest)
		assertProblemResponse(t, resp, map[string]any{
			"type":   "https://httpstatuses.com/400",
			"title":  "Invalid Query Parameter",
			"status": float64(400),
			"detail": "limit and min_similarity must be valid numbers",
			"path":   "/products/1/neighbors",
		})
	})

	t.Run("reject missing source embedding", func(t *testing.T) {
		repo := newMemoryProductRepository()
		repo.nextID = 2
		repo.products[1] = domainproduct.Product{
			ID:          1,
			CompanyID:   1,
			BranchID:    1,
			SKU:         "SEED-PRD-001",
			Name:        "Wireless Mouse",
			Description: "Mouse inalambrico",
			Category:    "peripherals",
			Brand:       "Acme",
			PriceCents:  19990,
			Currency:    "CLP",
			Stock:       10,
		}

		handler := producthttp.NewHandler(productapp.NewUseCase(repo, testProductEmbedder{}))
		server := router.New(nil, nil, nil, nil, handler, nil, nil)

		resp := performRequest(t, server, http.MethodGet, "/products/1/neighbors", "")
		assertStatus(t, resp, http.StatusUnprocessableEntity)
		assertProblemResponse(t, resp, map[string]any{
			"type":   "https://httpstatuses.com/422",
			"title":  "Validation Failed",
			"status": float64(422),
			"detail": "source product does not have embedding",
			"path":   "/products/1/neighbors",
		})
	})
}

type failingRepository struct {
	err error
}

type testProductEmbedder struct{}

func (testProductEmbedder) EmbedText(context.Context, string) ([]float32, error) {
	embedding := make([]float32, domainproduct.EmbeddingDimensions)
	for index := range embedding {
		embedding[index] = 0.001
	}

	return embedding, nil
}

func (r failingRepository) Create(context.Context, *domainproduct.Product) error  { return r.err }
func (r failingRepository) List(context.Context) ([]domainproduct.Product, error) { return nil, r.err }
func (r failingRepository) GetByID(context.Context, int64) (*domainproduct.Product, error) {
	return nil, r.err
}
func (r failingRepository) FindNeighbors(context.Context, int64, int64, int, float64) ([]productapp.NeighborOutput, error) {
	return nil, r.err
}
func (r failingRepository) Update(context.Context, *domainproduct.Product) error { return r.err }
func (r failingRepository) Delete(context.Context, int64) error                  { return r.err }

func mustCreateProduct(t *testing.T, handler http.Handler, body string) {
	t.Helper()

	resp := performRequest(t, handler, http.MethodPost, "/products", body)
	assertStatus(t, resp, http.StatusCreated)
}

func fakeDistance(sourceName, candidateName string) float64 {
	sourceTokens := strings.Fields(strings.ToLower(sourceName))
	candidateTokens := strings.Fields(strings.ToLower(candidateName))

	if len(sourceTokens) == 0 || len(candidateTokens) == 0 {
		return 1
	}

	matches := 0
	for _, source := range sourceTokens {
		for _, candidate := range candidateTokens {
			if source == candidate {
				matches++
				break
			}
		}
	}

	similarity := float64(matches) / float64(max(len(sourceTokens), len(candidateTokens)))
	return 1 - similarity
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func performRequest(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	return performRequestWithJSONOption(t, handler, method, path, body, body != "")
}

func performRequestWithJSONOption(t *testing.T, handler http.Handler, method, path, body string, setJSON bool) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if setJSON {
		request.Header.Set("Content-Type", "application/json")
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	return response
}

func assertStatus(t *testing.T, response *httptest.ResponseRecorder, want int) {
	t.Helper()

	if response.Code != want {
		t.Fatalf("expected status %d, got %d; body=%s", want, response.Code, response.Body.String())
	}
}

func assertHeader(t *testing.T, response *httptest.ResponseRecorder, key, want string) {
	t.Helper()

	if got := response.Header().Get(key); got != want {
		t.Fatalf("expected header %s=%q, got %q", key, want, got)
	}
}

func assertProblemResponse(t *testing.T, response *httptest.ResponseRecorder, want map[string]any) {
	t.Helper()

	if got := response.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("expected Content-Type application/problem+json, got %q", got)
	}

	var actual map[string]any
	decodeBody(t, response, &actual)
	assertJSONEqual(t, actual, want)
}

func decodeBody(t *testing.T, response *httptest.ResponseRecorder, dst any) {
	t.Helper()

	if err := json.Unmarshal(response.Body.Bytes(), dst); err != nil {
		t.Fatalf("unmarshal response body: %v", err)
	}
}

func assertJSONEqual(t *testing.T, actual, want map[string]any) {
	t.Helper()

	actualJSON, err := json.Marshal(actual)
	if err != nil {
		t.Fatalf("marshal actual JSON: %v", err)
	}

	wantJSON, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal expected JSON: %v", err)
	}

	if string(actualJSON) != string(wantJSON) {
		t.Fatalf("unexpected JSON response:\n got: %s\nwant: %s", actualJSON, wantJSON)
	}
}
