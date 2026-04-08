package producthttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"

	productapp "github.com/IanStuardo-Dev/backend-crud/internal/application/product"
	domainproduct "github.com/IanStuardo-Dev/backend-crud/internal/domain/product"
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
	var validationErr domainproduct.ValidationError

	switch {
	case errors.As(err, &validationErr):
		writeProblem(w, r, http.StatusUnprocessableEntity, "Validation Failed", "request validation failed", []invalidParam{{
			Field:  validationErr.Field,
			Reason: validationErr.Message,
		}})
	case errors.Is(err, productapp.ErrNotFound):
		writeProblem(w, r, http.StatusNotFound, "Resource Not Found", "product not found", nil)
	case errors.Is(err, productapp.ErrConflict):
		writeProblem(w, r, http.StatusConflict, "Conflict", "a product with the same sku already exists in this branch", nil)
	case errors.Is(err, productapp.ErrInvalidReference):
		writeProblem(w, r, http.StatusUnprocessableEntity, "Validation Failed", "company or branch reference is invalid", nil)
	case errors.Is(err, productapp.ErrSourceEmbeddingMissing):
		writeProblem(w, r, http.StatusUnprocessableEntity, "Validation Failed", "source product does not have embedding", nil)
	case errors.Is(err, productapp.ErrEmbeddingUnavailable):
		writeProblem(w, r, http.StatusServiceUnavailable, "Service Unavailable", "product embedding provider is not configured", nil)
	case errors.Is(err, productapp.ErrEmbeddingGeneration):
		writeProblem(w, r, http.StatusBadGateway, "Bad Gateway", "product embedding could not be generated", nil)
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

func requireJSONContentType(r *http.Request) error {
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		return errors.New("Content-Type must be application/json")
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "application/json" {
		return errors.New("Content-Type must be application/json")
	}

	return nil
}

func decodeJSONBody(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return errors.New("request body must be valid JSON")
	}
	if decoder.More() {
		return errors.New("request body must contain a single JSON object")
	}

	return nil
}
