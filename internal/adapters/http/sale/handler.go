package salehttp

import (
	"fmt"
	"net/http"
	"strconv"

	authhttp "github.com/example/crud/internal/adapters/http/auth"
	saleapp "github.com/example/crud/internal/application/sale"
	"github.com/gorilla/mux"
)

type Handler struct {
	useCase saleapp.UseCase
}

func NewHandler(useCase saleapp.UseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if err := requireJSONContentType(r); err != nil {
		writeProblem(w, r, http.StatusUnsupportedMediaType, "Unsupported Media Type", err.Error(), nil)
		return
	}

	var request createSaleRequest
	if err := decodeJSONBody(r, &request); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "Invalid Request Body", err.Error(), nil)
		return
	}

	authenticatedUser, ok := authhttp.AuthenticatedUserFromContext(r.Context())
	if !ok {
		writeProblem(w, r, http.StatusUnauthorized, "Unauthorized", "missing or invalid bearer token", nil)
		return
	}

	if err := authhttp.EnsureCompanyAccess(r.Context(), request.CompanyID); err != nil {
		writeProblem(w, r, http.StatusForbidden, "Forbidden", "you cannot access resources outside your company", nil)
		return
	}
	output, err := h.useCase.Create(r.Context(), toCreateInput(request, authenticatedUser.ID))
	if err != nil {
		writeApplicationError(w, r, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/sales/%d", output.ID))
	writeJSON(w, http.StatusCreated, resourceResponse{Data: toSaleResponse(output)})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	outputs, err := h.useCase.List(r.Context())
	if err != nil {
		writeApplicationError(w, r, err)
		return
	}
	if authenticatedUser, ok := authhttp.AuthenticatedUserFromContext(r.Context()); ok && !authenticatedUser.IsSuperAdmin() {
		outputs = filterSalesByCompany(outputs, *authenticatedUser.CompanyID)
	}

	responses := toSaleResponses(outputs)
	writeJSON(w, http.StatusOK, collectionResponse{
		Data: responses,
		Meta: metaResponse{Count: len(responses)},
	})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeProblem(w, r, http.StatusBadRequest, "Invalid Identifier", "id must be a valid integer", nil)
		return
	}

	output, err := h.useCase.GetByID(r.Context(), id)
	if err != nil {
		writeApplicationError(w, r, err)
		return
	}
	if err := authhttp.EnsureCompanyAccess(r.Context(), output.CompanyID); err != nil {
		writeProblem(w, r, http.StatusForbidden, "Forbidden", "you cannot access resources outside your company", nil)
		return
	}

	writeJSON(w, http.StatusOK, resourceResponse{Data: toSaleResponse(output)})
}

func parseIDParam(r *http.Request) (int64, error) {
	return strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
}

func filterSalesByCompany(outputs []saleapp.Output, companyID int64) []saleapp.Output {
	filtered := make([]saleapp.Output, 0, len(outputs))
	for _, output := range outputs {
		if output.CompanyID == companyID {
			filtered = append(filtered, output)
		}
	}
	return filtered
}
