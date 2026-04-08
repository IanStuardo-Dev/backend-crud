package transferhttp_test

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
	transferhttp "github.com/IanStuardo-Dev/backend-crud/internal/adapters/http/transfer"
	authapp "github.com/IanStuardo-Dev/backend-crud/internal/application/auth"
	transferapp "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer"
	domaintransfer "github.com/IanStuardo-Dev/backend-crud/internal/domain/transfer"
	"github.com/IanStuardo-Dev/backend-crud/internal/infrastructure/http/router"
)

type memoryTransferRepository struct {
	nextID    int64
	transfers map[int64]domaintransfer.Transfer
}

func newMemoryTransferRepository() *memoryTransferRepository {
	return &memoryTransferRepository{
		nextID:    1,
		transfers: make(map[int64]domaintransfer.Transfer),
	}
}

func (r *memoryTransferRepository) Create(_ context.Context, transfer *domaintransfer.Transfer) error {
	now := time.Now().UTC().Round(0)
	transfer.ID = r.nextID
	transfer.CreatedAt = now
	transfer.Status = domaintransfer.StatusPendingApproval
	r.nextID++
	for index := range transfer.Items {
		transfer.Items[index].ProductSKU = "SKU-010"
		transfer.Items[index].ProductName = "Mechanical Keyboard"
	}
	r.transfers[transfer.ID] = cloneTransfer(*transfer)
	return nil
}

func (r *memoryTransferRepository) Approve(_ context.Context, transfer *domaintransfer.Transfer) error {
	r.transfers[transfer.ID] = cloneTransfer(*transfer)
	return nil
}

func (r *memoryTransferRepository) Dispatch(_ context.Context, transfer *domaintransfer.Transfer) error {
	r.transfers[transfer.ID] = cloneTransfer(*transfer)
	return nil
}

func (r *memoryTransferRepository) Receive(_ context.Context, transfer *domaintransfer.Transfer) error {
	r.transfers[transfer.ID] = cloneTransfer(*transfer)
	return nil
}

func (r *memoryTransferRepository) Cancel(_ context.Context, transfer *domaintransfer.Transfer) error {
	r.transfers[transfer.ID] = cloneTransfer(*transfer)
	return nil
}

func (r *memoryTransferRepository) List(context.Context) ([]domaintransfer.Transfer, error) {
	transfers := make([]domaintransfer.Transfer, 0, len(r.transfers))
	for id := int64(1); id < r.nextID; id++ {
		if transfer, ok := r.transfers[id]; ok {
			transfers = append(transfers, cloneTransfer(transfer))
		}
	}
	return transfers, nil
}

func (r *memoryTransferRepository) ListByBranch(_ context.Context, branchID int64) ([]domaintransfer.Transfer, error) {
	transfers := make([]domaintransfer.Transfer, 0, len(r.transfers))
	for id := int64(1); id < r.nextID; id++ {
		if transfer, ok := r.transfers[id]; ok {
			if transfer.OriginBranchID == branchID || transfer.DestinationBranchID == branchID {
				transfers = append(transfers, cloneTransfer(transfer))
			}
		}
	}
	return transfers, nil
}

func (r *memoryTransferRepository) GetByID(_ context.Context, id int64) (*domaintransfer.Transfer, error) {
	transfer, ok := r.transfers[id]
	if !ok {
		return nil, nil
	}
	clone := cloneTransfer(transfer)
	return &clone, nil
}

func TestTransferWorkflowEndpoints(t *testing.T) {
	repo := newMemoryTransferRepository()
	handler := transferhttp.NewHandler(transferapp.NewUseCase(repo))
	server := router.New(nil, nil, authhttp.NewMiddleware(stubTokenVerifier{}), nil, nil, nil, handler)

	createBody := `{"company_id":1,"origin_branch_id":1,"destination_branch_id":2,"supervisor_user_id":9,"note":"restock north branch","items":[{"product_id":10,"quantity":2}]}`
	createResp := performAuthenticatedRequest(t, server, http.MethodPost, "/inventory/transfers", createBody, "manager-token")
	assertStatus(t, createResp, http.StatusCreated)
	assertHeader(t, createResp, "Location", "/inventory/transfers/1")

	approveResp := performAuthenticatedRequest(t, server, http.MethodPost, "/inventory/transfers/1/approve", "", "supervisor-token")
	assertStatus(t, approveResp, http.StatusOK)

	dispatchResp := performAuthenticatedRequest(t, server, http.MethodPost, "/inventory/transfers/1/dispatch", "", "manager-token")
	assertStatus(t, dispatchResp, http.StatusOK)

	receiveResp := performAuthenticatedRequest(t, server, http.MethodPost, "/inventory/transfers/1/receive", "", "receiver-token")
	assertStatus(t, receiveResp, http.StatusOK)

	getByIDResp := performAuthenticatedRequest(t, server, http.MethodGet, "/inventory/transfers/1", "", "manager-token")
	assertStatus(t, getByIDResp, http.StatusOK)

	listResp := performAuthenticatedRequest(t, server, http.MethodGet, "/inventory/transfers", "", "manager-token")
	assertStatus(t, listResp, http.StatusOK)

	listByBranchResp := performAuthenticatedRequest(t, server, http.MethodGet, "/inventory/transfers/branches/2", "", "manager-token")
	assertStatus(t, listByBranchResp, http.StatusOK)
}

func TestTransferApproveRejectsNonSupervisor(t *testing.T) {
	repo := newMemoryTransferRepository()
	useCase := transferapp.NewUseCase(repo)
	handler := transferhttp.NewHandler(useCase)
	server := router.New(nil, nil, authhttp.NewMiddleware(stubTokenVerifier{}), nil, nil, nil, handler)

	if _, err := useCase.Create(context.Background(), transferapp.CreateInput{
		CompanyID:           1,
		OriginBranchID:      1,
		DestinationBranchID: 2,
		RequestedByUserID:   7,
		SupervisorUserID:    9,
		Items: []transferapp.CreateItemInput{{
			ProductID: 10,
			Quantity:  1,
		}},
	}); err != nil {
		t.Fatalf("seed transfer: %v", err)
	}

	resp := performAuthenticatedRequest(t, server, http.MethodPost, "/inventory/transfers/1/approve", "", "manager-token")
	assertStatus(t, resp, http.StatusForbidden)
	assertProblemResponse(t, resp, map[string]any{
		"type":   "https://httpstatuses.com/403",
		"title":  "Forbidden",
		"status": float64(403),
		"detail": "forbidden transfer action\nonly the assigned supervisor can approve this transfer",
		"path":   "/inventory/transfers/1/approve",
	})
}

func TestTransferCreateValidatesSupervisorInBody(t *testing.T) {
	repo := newMemoryTransferRepository()
	handler := transferhttp.NewHandler(transferapp.NewUseCase(repo))
	server := router.New(nil, nil, authhttp.NewMiddleware(stubTokenVerifier{}), nil, nil, nil, handler)

	body := `{"company_id":1,"origin_branch_id":1,"destination_branch_id":2,"supervisor_user_id":7,"items":[{"product_id":10,"quantity":1}]}`
	resp := performAuthenticatedRequest(t, server, http.MethodPost, "/inventory/transfers", body, "manager-token")
	assertStatus(t, resp, http.StatusUnprocessableEntity)
	assertProblemResponse(t, resp, map[string]any{
		"type":   "https://httpstatuses.com/422",
		"title":  "Validation Failed",
		"status": float64(422),
		"detail": "request validation failed",
		"path":   "/inventory/transfers",
		"errors": []any{
			map[string]any{
				"field":  "supervisor_user_id",
				"reason": "supervisor_user_id must be different from requested_by_user_id",
			},
		},
	})
}

func TestTransferCreateReturnsDetailedUnknownFieldError(t *testing.T) {
	repo := newMemoryTransferRepository()
	handler := transferhttp.NewHandler(transferapp.NewUseCase(repo))
	server := router.New(nil, nil, authhttp.NewMiddleware(stubTokenVerifier{}), nil, nil, nil, handler)

	body := `{"company_id":1,"origin_branch_id":1,"destination_branch_id":2,"supervisorUserId":9,"items":[{"product_id":10,"quantity":1}]}`
	resp := performAuthenticatedRequest(t, server, http.MethodPost, "/inventory/transfers", body, "manager-token")
	assertStatus(t, resp, http.StatusBadRequest)
	assertProblemResponse(t, resp, map[string]any{
		"type":   "https://httpstatuses.com/400",
		"title":  "Invalid Request Body",
		"status": float64(400),
		"detail": "request body contains unknown field \"supervisorUserId\"",
		"path":   "/inventory/transfers",
	})
}

type stubTokenVerifier struct{}

func (stubTokenVerifier) Verify(token string) (authapp.AuthenticatedUser, error) {
	switch token {
	case "manager-token":
		return authapp.AuthenticatedUser{ID: 7, CompanyID: int64Pointer(1), Role: "inventory_manager", IsActive: true}, nil
	case "supervisor-token":
		return authapp.AuthenticatedUser{ID: 9, CompanyID: int64Pointer(1), Role: "company_admin", IsActive: true}, nil
	case "receiver-token":
		return authapp.AuthenticatedUser{ID: 11, CompanyID: int64Pointer(1), Role: "inventory_manager", IsActive: true}, nil
	default:
		return authapp.AuthenticatedUser{}, authapp.ErrUnauthorized
	}
}

func cloneTransfer(transfer domaintransfer.Transfer) domaintransfer.Transfer {
	clone := transfer
	clone.Items = append([]domaintransfer.Item(nil), transfer.Items...)
	return clone
}

func int64Pointer(value int64) *int64 {
	return &value
}

func performAuthenticatedRequest(t *testing.T, handler http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	request.Header.Set("Authorization", "Bearer "+token)
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

type failingRepository struct {
	err error
}

func (r failingRepository) Create(context.Context, *domaintransfer.Transfer) error   { return r.err }
func (r failingRepository) Approve(context.Context, *domaintransfer.Transfer) error  { return r.err }
func (r failingRepository) Dispatch(context.Context, *domaintransfer.Transfer) error { return r.err }
func (r failingRepository) Receive(context.Context, *domaintransfer.Transfer) error  { return r.err }
func (r failingRepository) Cancel(context.Context, *domaintransfer.Transfer) error   { return r.err }
func (r failingRepository) List(context.Context) ([]domaintransfer.Transfer, error) {
	return nil, r.err
}
func (r failingRepository) ListByBranch(context.Context, int64) ([]domaintransfer.Transfer, error) {
	return nil, r.err
}
func (r failingRepository) GetByID(context.Context, int64) (*domaintransfer.Transfer, error) {
	return nil, r.err
}

func TestTransferHandlerMapsUnexpectedErrorsToInternalServerError(t *testing.T) {
	useCase := transferapp.NewUseCase(failingRepository{err: errors.New("boom")})
	server := router.New(nil, nil, authhttp.NewMiddleware(stubTokenVerifier{}), nil, nil, nil, transferhttp.NewHandler(useCase))

	resp := performAuthenticatedRequest(t, server, http.MethodGet, "/inventory/transfers", "", "manager-token")
	assertStatus(t, resp, http.StatusInternalServerError)
	assertProblemResponse(t, resp, map[string]any{
		"type":   "https://httpstatuses.com/500",
		"title":  "Internal Server Error",
		"status": float64(500),
		"detail": "an unexpected error occurred",
		"path":   "/inventory/transfers",
	})
}
