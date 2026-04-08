package userhttp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	userhttp "github.com/IanStuardo-Dev/backend-crud/internal/adapters/http/user"
	userapp "github.com/IanStuardo-Dev/backend-crud/internal/application/user"
	domainuser "github.com/IanStuardo-Dev/backend-crud/internal/domain/user"
	"github.com/IanStuardo-Dev/backend-crud/internal/infrastructure/http/router"
)

type memoryUserRepository struct {
	nextID int64
	users  map[int64]domainuser.User
}

func newMemoryUserRepository() *memoryUserRepository {
	return &memoryUserRepository{
		nextID: 1,
		users:  make(map[int64]domainuser.User),
	}
}

func (r *memoryUserRepository) Create(_ context.Context, user *domainuser.User) error {
	for _, existing := range r.users {
		if existing.Email == user.Email {
			return userapp.ErrConflict
		}
	}

	user.ID = r.nextID
	r.nextID++
	r.users[user.ID] = *user

	return nil
}

func (r *memoryUserRepository) List(context.Context) ([]domainuser.User, error) {
	users := make([]domainuser.User, 0, len(r.users))
	for id := int64(1); id < r.nextID; id++ {
		if user, ok := r.users[id]; ok {
			users = append(users, user)
		}
	}

	return users, nil
}

func (r *memoryUserRepository) GetByID(_ context.Context, id int64) (*domainuser.User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, nil
	}

	userCopy := user
	return &userCopy, nil
}

func (r *memoryUserRepository) GetByEmail(_ context.Context, email string) (*domainuser.User, error) {
	for _, user := range r.users {
		if user.Email == email {
			userCopy := user
			return &userCopy, nil
		}
	}

	return nil, nil
}

func (r *memoryUserRepository) Update(_ context.Context, user *domainuser.User) error {
	if _, ok := r.users[user.ID]; !ok {
		return userapp.ErrNotFound
	}
	for id, existing := range r.users {
		if id != user.ID && existing.Email == user.Email {
			return userapp.ErrConflict
		}
	}

	r.users[user.ID] = *user
	return nil
}

func (r *memoryUserRepository) Delete(_ context.Context, id int64) error {
	if _, ok := r.users[id]; !ok {
		return userapp.ErrNotFound
	}

	delete(r.users, id)
	return nil
}

type testHasher struct{}

func (testHasher) Hash(password string) (string, error) {
	return "hashed:" + password, nil
}

func TestUserCRUDFlow(t *testing.T) {
	repo := newMemoryUserRepository()
	handler := userhttp.NewHandler(userapp.NewUseCase(repo, testHasher{}))
	server := router.New(handler, nil, nil, nil, nil, nil, nil)

	createBody := `{"company_id":1,"name":"  Alice Doe  ","email":"ALICE@example.com","role":"company_admin","is_active":true,"default_branch_id":1,"password":"Password123"}`
	createResp := performRequest(t, server, http.MethodPost, "/users", createBody)
	assertStatus(t, createResp, http.StatusCreated)
	assertHeader(t, createResp, "Location", "/users/1")
	assertJSONResponse(t, createResp, map[string]any{
		"data": map[string]any{
			"id":                float64(1),
			"company_id":        float64(1),
			"name":              "Alice Doe",
			"email":             "alice@example.com",
			"role":              "company_admin",
			"is_active":         true,
			"default_branch_id": float64(1),
		},
	})

	getAllResp := performRequest(t, server, http.MethodGet, "/users", "")
	assertStatus(t, getAllResp, http.StatusOK)
	assertJSONResponse(t, getAllResp, map[string]any{
		"data": []any{
			map[string]any{
				"id":                float64(1),
				"company_id":        float64(1),
				"name":              "Alice Doe",
				"email":             "alice@example.com",
				"role":              "company_admin",
				"is_active":         true,
				"default_branch_id": float64(1),
			},
		},
		"meta": map[string]any{
			"count": float64(1),
		},
	})

	getByIDResp := performRequest(t, server, http.MethodGet, "/users/1", "")
	assertStatus(t, getByIDResp, http.StatusOK)
	assertJSONResponse(t, getByIDResp, map[string]any{
		"data": map[string]any{
			"id":                float64(1),
			"company_id":        float64(1),
			"name":              "Alice Doe",
			"email":             "alice@example.com",
			"role":              "company_admin",
			"is_active":         true,
			"default_branch_id": float64(1),
		},
	})

	updateBody := `{"company_id":1,"name":"Alice Smith","email":"alice.smith@example.com","role":"company_admin","is_active":true,"default_branch_id":1}`
	updateResp := performRequest(t, server, http.MethodPut, "/users/1", updateBody)
	assertStatus(t, updateResp, http.StatusOK)
	assertJSONResponse(t, updateResp, map[string]any{
		"data": map[string]any{
			"id":                float64(1),
			"company_id":        float64(1),
			"name":              "Alice Smith",
			"email":             "alice.smith@example.com",
			"role":              "company_admin",
			"is_active":         true,
			"default_branch_id": float64(1),
		},
	})

	deleteResp := performRequest(t, server, http.MethodDelete, "/users/1", "")
	assertStatus(t, deleteResp, http.StatusNoContent)

	notFoundResp := performRequest(t, server, http.MethodGet, "/users/1", "")
	assertStatus(t, notFoundResp, http.StatusNotFound)
	assertProblemResponse(t, notFoundResp, map[string]any{
		"type":   "https://httpstatuses.com/404",
		"title":  "Resource Not Found",
		"status": float64(404),
		"detail": "user not found",
		"path":   "/users/1",
	})
}

func TestUserCRUDValidationAndErrors(t *testing.T) {
	repo := newMemoryUserRepository()
	handler := userhttp.NewHandler(userapp.NewUseCase(repo, testHasher{}))
	server := router.New(handler, nil, nil, nil, nil, nil, nil)

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
			name:       "reject invalid json on create",
			method:     http.MethodPost,
			path:       "/users",
			body:       `{"name":`,
			setJSONCT:  true,
			statusCode: http.StatusBadRequest,
			wantBody: map[string]any{
				"type":   "https://httpstatuses.com/400",
				"title":  "Invalid Request Body",
				"status": float64(400),
				"detail": "request body must be valid JSON",
				"path":   "/users",
			},
		},
		{
			name:       "reject empty name",
			method:     http.MethodPost,
			path:       "/users",
			body:       `{"company_id":1,"name":" ","email":"alice@example.com","role":"company_admin","is_active":true,"default_branch_id":1,"password":"Password123"}`,
			setJSONCT:  true,
			statusCode: http.StatusUnprocessableEntity,
			wantBody: map[string]any{
				"type":   "https://httpstatuses.com/422",
				"title":  "Validation Failed",
				"status": float64(422),
				"detail": "request validation failed",
				"path":   "/users",
				"errors": []any{
					map[string]any{
						"field":  "name",
						"reason": "name is required",
					},
				},
			},
		},
		{
			name:       "reject short name",
			method:     http.MethodPost,
			path:       "/users",
			body:       `{"company_id":1,"name":"A","email":"alice@example.com","role":"company_admin","is_active":true,"default_branch_id":1,"password":"Password123"}`,
			setJSONCT:  true,
			statusCode: http.StatusUnprocessableEntity,
			wantBody: map[string]any{
				"type":   "https://httpstatuses.com/422",
				"title":  "Validation Failed",
				"status": float64(422),
				"detail": "request validation failed",
				"path":   "/users",
				"errors": []any{
					map[string]any{
						"field":  "name",
						"reason": "name must be at least 2 characters",
					},
				},
			},
		},
		{
			name:       "reject invalid email",
			method:     http.MethodPost,
			path:       "/users",
			body:       `{"company_id":1,"name":"Alice","email":"alice-at-example.com","role":"company_admin","is_active":true,"default_branch_id":1,"password":"Password123"}`,
			setJSONCT:  true,
			statusCode: http.StatusUnprocessableEntity,
			wantBody: map[string]any{
				"type":   "https://httpstatuses.com/422",
				"title":  "Validation Failed",
				"status": float64(422),
				"detail": "request validation failed",
				"path":   "/users",
				"errors": []any{
					map[string]any{
						"field":  "email",
						"reason": "email format is invalid",
					},
				},
			},
		},
		{
			name:       "reject duplicate email on create",
			method:     http.MethodPost,
			path:       "/users",
			body:       `{"company_id":1,"name":"Bob","email":"alice@example.com","role":"company_admin","is_active":true,"default_branch_id":1,"password":"Password123"}`,
			setJSONCT:  true,
			statusCode: http.StatusConflict,
			wantBody: map[string]any{
				"type":   "https://httpstatuses.com/409",
				"title":  "Conflict",
				"status": float64(409),
				"detail": "a user with the same email already exists",
				"path":   "/users",
			},
			prepare: func(t *testing.T) {
				t.Helper()
				err := repo.Create(context.Background(), &domainuser.User{CompanyID: int64Pointer(1), Name: "Alice", Email: "alice@example.com", Role: domainuser.RoleCompanyAdmin, IsActive: true, DefaultBranchID: int64Pointer(1), PasswordHash: "hashed:Password123"})
				if err != nil {
					t.Fatalf("seed user: %v", err)
				}
			},
		},
		{
			name:       "reject weak password",
			method:     http.MethodPost,
			path:       "/users",
			body:       `{"company_id":1,"name":"Alice","email":"alice@example.com","role":"company_admin","is_active":true,"default_branch_id":1,"password":"short"}`,
			setJSONCT:  true,
			statusCode: http.StatusUnprocessableEntity,
			wantBody: map[string]any{
				"type":   "https://httpstatuses.com/422",
				"title":  "Validation Failed",
				"status": float64(422),
				"detail": "request validation failed",
				"path":   "/users",
				"errors": []any{
					map[string]any{
						"field":  "password",
						"reason": "password must be at least 8 characters",
					},
				},
			},
		},
		{
			name:       "reject missing content type",
			method:     http.MethodPost,
			path:       "/users",
			body:       `{"company_id":1,"name":"Alice","email":"alice@example.com","role":"company_admin","is_active":true,"default_branch_id":1,"password":"Password123"}`,
			statusCode: http.StatusUnsupportedMediaType,
			wantBody: map[string]any{
				"type":   "https://httpstatuses.com/415",
				"title":  "Unsupported Media Type",
				"status": float64(415),
				"detail": "Content-Type must be application/json",
				"path":   "/users",
			},
		},
		{
			name:       "reject invalid identifier",
			method:     http.MethodDelete,
			path:       "/users/abc",
			statusCode: http.StatusNotFound,
			wantBody: map[string]any{
				"type":   "https://httpstatuses.com/404",
				"title":  "Resource Not Found",
				"status": float64(404),
				"detail": "endpoint not found",
				"path":   "/users/abc",
			},
		},
		{
			name:       "reject not found on update",
			method:     http.MethodPut,
			path:       "/users/20",
			body:       `{"company_id":1,"name":"Alice","email":"alice@example.com","role":"company_admin","is_active":true,"default_branch_id":1}`,
			setJSONCT:  true,
			statusCode: http.StatusNotFound,
			wantBody: map[string]any{
				"type":   "https://httpstatuses.com/404",
				"title":  "Resource Not Found",
				"status": float64(404),
				"detail": "user not found",
				"path":   "/users/20",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo.nextID = 1
			repo.users = make(map[int64]domainuser.User)
			if tc.prepare != nil {
				tc.prepare(t)
			}

			resp := performRequestWithJSONOption(t, server, tc.method, tc.path, tc.body, tc.setJSONCT)
			assertStatus(t, resp, tc.statusCode)

			if tc.statusCode == http.StatusNoContent {
				return
			}

			assertProblemResponse(t, resp, tc.wantBody)
		})
	}
}

func TestHandlerMapsUnexpectedErrorsToInternalServerError(t *testing.T) {
	useCase := userapp.NewUseCase(failingRepository{err: errors.New("boom")}, testHasher{})
	server := router.New(userhttp.NewHandler(useCase), nil, nil, nil, nil, nil, nil)

	resp := performRequest(t, server, http.MethodGet, "/users", "")

	assertStatus(t, resp, http.StatusInternalServerError)
	assertProblemResponse(t, resp, map[string]any{
		"type":   "https://httpstatuses.com/500",
		"title":  "Internal Server Error",
		"status": float64(500),
		"detail": "an unexpected error occurred",
		"path":   "/users",
	})
}

type failingRepository struct {
	err error
}

func (r failingRepository) Create(context.Context, *domainuser.User) error  { return r.err }
func (r failingRepository) List(context.Context) ([]domainuser.User, error) { return nil, r.err }
func (r failingRepository) GetByID(context.Context, int64) (*domainuser.User, error) {
	return nil, r.err
}
func (r failingRepository) GetByEmail(context.Context, string) (*domainuser.User, error) {
	return nil, r.err
}
func (r failingRepository) Update(context.Context, *domainuser.User) error { return r.err }
func (r failingRepository) Delete(context.Context, int64) error            { return r.err }

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

func assertJSONResponse(t *testing.T, response *httptest.ResponseRecorder, want map[string]any) {
	t.Helper()

	if got := response.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", got)
	}

	var actual map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &actual); err != nil {
		t.Fatalf("unmarshal response body: %v", err)
	}

	assertJSONEqual(t, actual, want)
}

func assertProblemResponse(t *testing.T, response *httptest.ResponseRecorder, want map[string]any) {
	t.Helper()

	if got := response.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("expected Content-Type application/problem+json, got %q", got)
	}

	var actual map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &actual); err != nil {
		t.Fatalf("unmarshal problem response: %v", err)
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

func int64Pointer(value int64) *int64 {
	return &value
}
