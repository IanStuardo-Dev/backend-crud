package salehttp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authhttp "github.com/IanStuardo-Dev/backend-crud/internal/adapters/http/auth"
	salehttp "github.com/IanStuardo-Dev/backend-crud/internal/adapters/http/sale"
	authapp "github.com/IanStuardo-Dev/backend-crud/internal/application/auth"
	saleapp "github.com/IanStuardo-Dev/backend-crud/internal/application/sale"
	domainsale "github.com/IanStuardo-Dev/backend-crud/internal/domain/sale"
	"github.com/IanStuardo-Dev/backend-crud/internal/infrastructure/http/router"
)

type memorySaleRepository struct {
	nextID int64
	sales  map[int64]domainsale.Sale
}

func newMemorySaleRepository() *memorySaleRepository {
	return &memorySaleRepository{
		nextID: 1,
		sales:  make(map[int64]domainsale.Sale),
	}
}

func (r *memorySaleRepository) Create(_ context.Context, sale *domainsale.Sale) error {
	now := time.Now().UTC().Round(0)
	sale.ID = r.nextID
	sale.CreatedAt = now
	r.nextID++

	var total int64
	for index := range sale.Items {
		sale.Items[index].ProductSKU = "SKU-001"
		sale.Items[index].ProductName = "Laptop Stand"
		sale.Items[index].UnitPriceCents = 5000
		sale.Items[index].SubtotalCents = int64(sale.Items[index].Quantity) * sale.Items[index].UnitPriceCents
		total += sale.Items[index].SubtotalCents
	}
	sale.TotalAmountCents = total
	r.sales[sale.ID] = *sale
	return nil
}

func (r *memorySaleRepository) List(context.Context) ([]domainsale.Sale, error) {
	sales := make([]domainsale.Sale, 0, len(r.sales))
	for id := int64(1); id < r.nextID; id++ {
		if sale, ok := r.sales[id]; ok {
			sales = append(sales, sale)
		}
	}
	return sales, nil
}

func (r *memorySaleRepository) GetByID(_ context.Context, id int64) (*domainsale.Sale, error) {
	sale, ok := r.sales[id]
	if !ok {
		return nil, nil
	}
	saleCopy := sale
	return &saleCopy, nil
}

func (r *memorySaleRepository) BranchExists(context.Context, int64, int64) (bool, error) {
	return true, nil
}

func (r *memorySaleRepository) UserExists(context.Context, int64) (bool, error) {
	return true, nil
}

func (r *memorySaleRepository) LoadForSale(_ context.Context, companyID, branchID int64, items []domainsale.Item, lock bool) (map[int64]saleapp.StockSnapshot, error) {
	products := make(map[int64]saleapp.StockSnapshot, len(items))
	for _, item := range items {
		products[item.ProductID] = saleapp.StockSnapshot{
			ProductID:      item.ProductID,
			SKU:            "SKU-001",
			Name:           "Laptop Stand",
			PriceCents:     5000,
			StockOnHand:    10,
			ReservedStock:  0,
			AvailableStock: 10,
		}
	}
	return products, nil
}

func (r *memorySaleRepository) SetStockOnHand(context.Context, int64, int64, int64, int) error {
	return nil
}

func (r *memorySaleRepository) CreateSaleMovement(context.Context, saleapp.MovementInput) error {
	return nil
}

func TestSaleCreateAndReadFlow(t *testing.T) {
	repo := newMemorySaleRepository()
	handler := salehttp.NewHandler(saleapp.NewUseCase(repo))
	server := router.New(nil, nil, authhttp.NewMiddleware(stubTokenVerifier{}), nil, nil, handler, nil)

	createBody := `{"company_id":1,"branch_id":1,"items":[{"product_id":9,"quantity":2}]}`
	createResp := performAuthenticatedRequest(t, server, http.MethodPost, "/sales", createBody)
	assertStatus(t, createResp, http.StatusCreated)
	assertHeader(t, createResp, "Location", "/sales/1")

	getAllResp := performAuthenticatedRequest(t, server, http.MethodGet, "/sales", "")
	assertStatus(t, getAllResp, http.StatusOK)

	getByIDResp := performAuthenticatedRequest(t, server, http.MethodGet, "/sales/1", "")
	assertStatus(t, getByIDResp, http.StatusOK)
}

func TestSaleValidationAndErrors(t *testing.T) {
	repo := newMemorySaleRepository()
	handler := salehttp.NewHandler(saleapp.NewUseCase(repo))
	server := router.New(nil, nil, authhttp.NewMiddleware(stubTokenVerifier{}), nil, nil, handler, nil)

	cases := []struct {
		name       string
		method     string
		path       string
		body       string
		setJSONCT  bool
		setAuth    bool
		statusCode int
		wantBody   map[string]any
	}{
		{
			name:       "reject duplicate product item",
			method:     http.MethodPost,
			path:       "/sales",
			body:       `{"company_id":1,"branch_id":1,"items":[{"product_id":9,"quantity":1},{"product_id":9,"quantity":2}]}`,
			setJSONCT:  true,
			setAuth:    true,
			statusCode: http.StatusUnprocessableEntity,
			wantBody: map[string]any{
				"type":   "https://httpstatuses.com/422",
				"title":  "Validation Failed",
				"status": float64(422),
				"detail": "request validation failed",
				"path":   "/sales",
				"errors": []any{
					map[string]any{
						"field":  "items",
						"reason": "items must not repeat the same product",
					},
				},
			},
		},
		{
			name:       "reject missing content type",
			method:     http.MethodPost,
			path:       "/sales",
			body:       `{"company_id":1,"branch_id":1,"items":[{"product_id":9,"quantity":1}]}`,
			setJSONCT:  false,
			setAuth:    true,
			statusCode: http.StatusUnsupportedMediaType,
			wantBody: map[string]any{
				"type":   "https://httpstatuses.com/415",
				"title":  "Unsupported Media Type",
				"status": float64(415),
				"detail": "Content-Type must be application/json",
				"path":   "/sales",
			},
		},
		{
			name:       "reject unauthenticated sale creation",
			method:     http.MethodPost,
			path:       "/sales",
			body:       `{"company_id":1,"branch_id":1,"items":[{"product_id":9,"quantity":1}]}`,
			setJSONCT:  true,
			setAuth:    false,
			statusCode: http.StatusUnauthorized,
			wantBody: map[string]any{
				"type":   "https://httpstatuses.com/401",
				"title":  "Unauthorized",
				"status": float64(401),
				"detail": "missing or invalid bearer token",
				"path":   "/sales",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := performRequestWithJSONAndAuthOption(t, server, tc.method, tc.path, tc.body, tc.setJSONCT, tc.setAuth)
			assertStatus(t, resp, tc.statusCode)
			assertProblemResponse(t, resp, tc.wantBody)
		})
	}
}

func TestSaleHandlerMapsUnexpectedErrorsToInternalServerError(t *testing.T) {
	useCase := saleapp.NewUseCase(failingRepository{err: errors.New("boom")})
	server := router.New(nil, nil, nil, nil, nil, salehttp.NewHandler(useCase), nil)

	resp := performRequest(t, server, http.MethodGet, "/sales", "")
	assertStatus(t, resp, http.StatusInternalServerError)
	assertProblemResponse(t, resp, map[string]any{
		"type":   "https://httpstatuses.com/500",
		"title":  "Internal Server Error",
		"status": float64(500),
		"detail": "an unexpected error occurred",
		"path":   "/sales",
	})
}

type failingRepository struct {
	err error
}

func (r failingRepository) Create(context.Context, *domainsale.Sale) error  { return r.err }
func (r failingRepository) List(context.Context) ([]domainsale.Sale, error) { return nil, r.err }
func (r failingRepository) GetByID(context.Context, int64) (*domainsale.Sale, error) {
	return nil, r.err
}
func (r failingRepository) BranchExists(context.Context, int64, int64) (bool, error) {
	return false, r.err
}
func (r failingRepository) UserExists(context.Context, int64) (bool, error) { return false, r.err }
func (r failingRepository) LoadForSale(context.Context, int64, int64, []domainsale.Item, bool) (map[int64]saleapp.StockSnapshot, error) {
	return nil, r.err
}
func (r failingRepository) SetStockOnHand(context.Context, int64, int64, int64, int) error {
	return r.err
}
func (r failingRepository) CreateSaleMovement(context.Context, saleapp.MovementInput) error {
	return r.err
}

type stubTokenVerifier struct{}

func (stubTokenVerifier) Verify(token string) (authapp.AuthenticatedUser, error) {
	if token != "valid-token" {
		return authapp.AuthenticatedUser{}, authapp.ErrUnauthorized
	}

	return authapp.AuthenticatedUser{
		ID:        7,
		CompanyID: int64Pointer(1),
		Name:      "Operator",
		Email:     "operator@example.com",
		Role:      "sales_user",
		IsActive:  true,
	}, nil
}

func int64Pointer(value int64) *int64 {
	return &value
}

func performRequest(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	return performRequestWithJSONOption(t, handler, method, path, body, body != "")
}

func performAuthenticatedRequest(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	return performRequestWithJSONAndAuthOption(t, handler, method, path, body, body != "", true)
}

func performRequestWithJSONOption(t *testing.T, handler http.Handler, method, path, body string, setJSON bool) *httptest.ResponseRecorder {
	t.Helper()
	return performRequestWithJSONAndAuthOption(t, handler, method, path, body, setJSON, false)
}

func performRequestWithJSONAndAuthOption(t *testing.T, handler http.Handler, method, path, body string, setJSON bool, setAuth bool) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if setJSON {
		request.Header.Set("Content-Type", "application/json")
	}
	if setAuth {
		request.Header.Set("Authorization", "Bearer valid-token")
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
	if err := json.Unmarshal(response.Body.Bytes(), &actual); err != nil {
		t.Fatalf("unmarshal response body: %v", err)
	}
	assertJSONEqual(t, actual, want)
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
