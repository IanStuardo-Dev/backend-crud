package transferhttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"

	transferapp "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer"
	domaintransfer "github.com/IanStuardo-Dev/backend-crud/internal/domain/transfer"
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
	var validationErr domaintransfer.ValidationError

	switch {
	case errors.As(err, &validationErr):
		writeProblem(w, r, http.StatusUnprocessableEntity, "Validation Failed", "request validation failed", []invalidParam{{
			Field:  validationErr.Field,
			Reason: validationErr.Message,
		}})
	case errors.Is(err, transferapp.ErrNotFound):
		writeProblem(w, r, http.StatusNotFound, "Resource Not Found", "transfer not found", nil)
	case errors.Is(err, transferapp.ErrForbiddenAction):
		writeProblem(w, r, http.StatusForbidden, "Forbidden", err.Error(), nil)
	case errors.Is(err, transferapp.ErrInvalidReference):
		writeProblem(w, r, http.StatusUnprocessableEntity, "Validation Failed", "company, branch, user, or product reference is invalid", nil)
	case errors.Is(err, transferapp.ErrInvalidState):
		writeProblem(w, r, http.StatusConflict, "Conflict", err.Error(), nil)
	case errors.Is(err, transferapp.ErrInsufficientStock):
		writeProblem(w, r, http.StatusConflict, "Conflict", "origin branch does not have enough available stock", nil)
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
