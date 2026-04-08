package authhttp

import (
	"net/http"
	"strings"

	authapp "github.com/example/crud/internal/application/auth"
)

type tokenVerifier interface {
	Verify(token string) (authapp.AuthenticatedUser, error)
}

type Middleware struct {
	tokens tokenVerifier
}

func NewMiddleware(tokens tokenVerifier) *Middleware {
	return &Middleware{tokens: tokens}
}

func (m *Middleware) RequireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			writeProblem(w, r, http.StatusUnauthorized, "Unauthorized", "missing or invalid bearer token", nil)
			return
		}

		user, err := m.tokens.Verify(token)
		if err != nil {
			writeProblem(w, r, http.StatusUnauthorized, "Unauthorized", "missing or invalid bearer token", nil)
			return
		}
		if !user.IsActive {
			writeProblem(w, r, http.StatusForbidden, "Forbidden", "user account is inactive", nil)
			return
		}

		next.ServeHTTP(w, r.WithContext(withAuthenticatedUser(r.Context(), user)))
	})
}

func (m *Middleware) RequireRoles(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := AuthenticatedUserFromContext(r.Context())
			if !ok {
				writeProblem(w, r, http.StatusUnauthorized, "Unauthorized", "missing or invalid bearer token", nil)
				return
			}
			if user.IsSuperAdmin() || user.HasAnyRole(roles...) {
				next.ServeHTTP(w, r)
				return
			}

			writeProblem(w, r, http.StatusForbidden, "Forbidden", "you do not have permission to perform this action", nil)
		})
	}
}

func bearerToken(header string) (string, bool) {
	if header == "" {
		return "", false
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", false
	}

	return token, true
}
