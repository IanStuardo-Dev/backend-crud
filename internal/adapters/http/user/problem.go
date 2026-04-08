package userhttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"

	authapp "github.com/IanStuardo-Dev/backend-crud/internal/application/auth"
	userapp "github.com/IanStuardo-Dev/backend-crud/internal/application/user"
	domainuser "github.com/IanStuardo-Dev/backend-crud/internal/domain/user"
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
	var validationErr domainuser.ValidationError

	switch {
	case errors.As(err, &validationErr):
		writeProblem(w, r, http.StatusUnprocessableEntity, "Validation Failed", "request validation failed", []invalidParam{{
			Field:  validationErr.Field,
			Reason: validationErr.Message,
		}})
	case errors.Is(err, userapp.ErrNotFound):
		writeProblem(w, r, http.StatusNotFound, "Resource Not Found", "user not found", nil)
	case errors.Is(err, userapp.ErrConflict):
		writeProblem(w, r, http.StatusConflict, "Conflict", "a user with the same email already exists", nil)
	case errors.Is(err, authapp.ErrForbidden):
		writeProblem(w, r, http.StatusForbidden, "Forbidden", "you do not have permission to perform this action", nil)
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
