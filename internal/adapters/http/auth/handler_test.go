package authhttp_test

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
	authapp "github.com/IanStuardo-Dev/backend-crud/internal/application/auth"
	"github.com/IanStuardo-Dev/backend-crud/internal/infrastructure/http/router"
)

type stubAuthUseCase struct {
	loginFn func(context.Context, authapp.LoginInput) (authapp.LoginOutput, error)
}

func (s stubAuthUseCase) Login(ctx context.Context, input authapp.LoginInput) (authapp.LoginOutput, error) {
	if s.loginFn != nil {
		return s.loginFn(ctx, input)
	}

	return authapp.LoginOutput{}, nil
}

func TestHandlerLoginReturnsToken(t *testing.T) {
	expiresAt := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)
	handler := authhttp.NewHandler(stubAuthUseCase{
		loginFn: func(_ context.Context, input authapp.LoginInput) (authapp.LoginOutput, error) {
			if input.Email != "alice@example.com" {
				t.Fatalf("expected email alice@example.com, got %q", input.Email)
			}
			return authapp.LoginOutput{
				AccessToken: "token-123",
				TokenType:   "Bearer",
				ExpiresAt:   expiresAt,
				User: authapp.UserOutput{
					ID:    1,
					Name:  "Alice",
					Email: "alice@example.com",
				},
			}, nil
		},
	})
	server := router.New(nil, handler, nil, nil, nil, nil, nil)

	request := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"alice@example.com","password":"Password123"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", response.Code, response.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	data := body["data"].(map[string]any)
	if data["access_token"] != "token-123" {
		t.Fatalf("expected access token token-123, got %#v", data["access_token"])
	}
}

func TestHandlerLoginMapsInvalidCredentials(t *testing.T) {
	handler := authhttp.NewHandler(stubAuthUseCase{
		loginFn: func(context.Context, authapp.LoginInput) (authapp.LoginOutput, error) {
			return authapp.LoginOutput{}, authapp.ErrInvalidCredentials
		},
	})
	server := router.New(nil, handler, nil, nil, nil, nil, nil)

	request := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"alice@example.com","password":"bad"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d body=%s", response.Code, response.Body.String())
	}
}

func TestHandlerLoginMapsUnexpectedErrors(t *testing.T) {
	handler := authhttp.NewHandler(stubAuthUseCase{
		loginFn: func(context.Context, authapp.LoginInput) (authapp.LoginOutput, error) {
			return authapp.LoginOutput{}, errors.New("boom")
		},
	})
	server := router.New(nil, handler, nil, nil, nil, nil, nil)

	request := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"alice@example.com","password":"Password123"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d body=%s", response.Code, response.Body.String())
	}
}
