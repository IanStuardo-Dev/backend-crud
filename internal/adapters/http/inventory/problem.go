package inventoryhttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	inventoryapp "github.com/IanStuardo-Dev/backend-crud/internal/application/inventory"
	domaininventory "github.com/IanStuardo-Dev/backend-crud/internal/domain/inventory"
)

type problemDetail struct {
	Type   string         `json:"type"`
	Title  string         `json:"title"`
	Status int            `json:"status"`
	Detail string         `json:"detail"`
	Path   string         `json:"path,omitempty"`
	Errors []invalidParam `json:"errors,omitempty"`
}

type invalidParam struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

func writeApplicationError(w http.ResponseWriter, r *http.Request, err error) {
	var validationErr domaininventory.ValidationError

	switch {
	case errors.As(err, &validationErr):
		writeProblem(w, r, http.StatusUnprocessableEntity, "Validation Failed", "request validation failed", []invalidParam{{
			Field:  validationErr.Field,
			Reason: validationErr.Message,
		}})
	case errors.Is(err, inventoryapp.ErrInvalidReference):
		writeProblem(w, r, http.StatusUnprocessableEntity, "Validation Failed", "company, branch, or product reference is invalid", nil)
	default:
		writeProblem(w, r, http.StatusInternalServerError, "Internal Server Error", "an unexpected error occurred", nil)
	}
}

func writeProblem(w http.ResponseWriter, r *http.Request, status int, title, detail string, invalidParams []invalidParam) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(problemDetail{
		Type:   problemTypeURI(status),
		Title:  title,
		Status: status,
		Detail: detail,
		Path:   r.URL.Path,
		Errors: invalidParams,
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func problemTypeURI(status int) string {
	return fmt.Sprintf("https://httpstatuses.com/%d", status)
}
