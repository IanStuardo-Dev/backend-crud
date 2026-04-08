package authhttp_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	authhttp "github.com/IanStuardo-Dev/backend-crud/internal/adapters/http/auth"
	authapp "github.com/IanStuardo-Dev/backend-crud/internal/application/auth"
)

type stubVerifier struct {
	verifyFn func(string) (authapp.AuthenticatedUser, error)
}

func (s stubVerifier) Verify(token string) (authapp.AuthenticatedUser, error) {
	if s.verifyFn != nil {
		return s.verifyFn(token)
	}

	return authapp.AuthenticatedUser{}, nil
}

func TestMiddlewareRejectsMissingBearerToken(t *testing.T) {
	middleware := authhttp.NewMiddleware(stubVerifier{})
	handler := middleware.RequireAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not run")
	}))

	request := httptest.NewRequest(http.MethodPost, "/sales", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d body=%s", response.Code, response.Body.String())
	}
}

func TestMiddlewareInjectsAuthenticatedUserIntoContext(t *testing.T) {
	middleware := authhttp.NewMiddleware(stubVerifier{
		verifyFn: func(token string) (authapp.AuthenticatedUser, error) {
			if token != "valid-token" {
				return authapp.AuthenticatedUser{}, errors.New("unexpected token")
			}
			return authapp.AuthenticatedUser{
				ID:        9,
				CompanyID: int64Pointer(1),
				Name:      "Operator",
				Email:     "operator@example.com",
				Role:      "sales_user",
				IsActive:  true,
			}, nil
		},
	})

	handler := middleware.RequireAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := authhttp.AuthenticatedUserFromContext(r.Context())
		if !ok {
			t.Fatal("expected authenticated user in context")
		}
		if user.ID != 9 {
			t.Fatalf("expected user ID 9, got %d", user.ID)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodPost, "/sales", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d body=%s", response.Code, response.Body.String())
	}
}

func int64Pointer(value int64) *int64 {
	return &value
}
