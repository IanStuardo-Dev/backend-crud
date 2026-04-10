package producthttp

import (
	"fmt"
	"net/http"
	"strconv"

	authhttp "github.com/IanStuardo-Dev/backend-crud/internal/adapters/http/auth"
	productapp "github.com/IanStuardo-Dev/backend-crud/internal/application/product"
	"github.com/gorilla/mux"
)

type Handler struct {
	useCase productapp.UseCase
}

func NewHandler(useCase productapp.UseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if err := requireJSONContentType(r); err != nil {
		writeProblem(w, r, http.StatusUnsupportedMediaType, "Unsupported Media Type", err.Error(), nil)
		return
	}

	var request createProductRequest
	if err := decodeJSONBody(r, &request); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "Invalid Request Body", err.Error(), nil)
		return
	}
	if err := authhttp.EnsureCompanyAccess(r.Context(), request.CompanyID); err != nil {
		writeProblem(w, r, http.StatusForbidden, "Forbidden", "you cannot access resources outside your company", nil)
		return
	}

	output, err := h.useCase.Create(r.Context(), toCreateInput(request))
	if err != nil {
		writeApplicationError(w, r, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/products/%d", output.ID))
	writeJSON(w, http.StatusCreated, resourceResponse{Data: toProductResponse(output)})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	outputs, err := h.useCase.List(r.Context())
	if err != nil {
		writeApplicationError(w, r, err)
		return
	}
	if authenticatedUser, ok := authhttp.AuthenticatedUserFromContext(r.Context()); ok && !authenticatedUser.IsSuperAdmin() {
		outputs = filterProductsByCompany(outputs, *authenticatedUser.CompanyID)
	}

	responses := toProductResponses(outputs)
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

	writeJSON(w, http.StatusOK, resourceResponse{Data: toProductResponse(output)})
}

func (h *Handler) FindNeighbors(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeProblem(w, r, http.StatusBadRequest, "Invalid Identifier", "id must be a valid integer", nil)
		return
	}

	input, err := toFindNeighborsInput(id, r.URL.Query().Get("limit"), r.URL.Query().Get("min_similarity"))
	if err != nil {
		writeProblem(w, r, http.StatusBadRequest, "Invalid Query Parameter", "limit and min_similarity must be valid numbers", nil)
		return
	}

	output, err := h.useCase.FindNeighbors(r.Context(), input)
	if err != nil {
		writeApplicationError(w, r, err)
		return
	}
	if err := authhttp.EnsureCompanyAccess(r.Context(), output.SourceCompanyID); err != nil {
		writeProblem(w, r, http.StatusForbidden, "Forbidden", "you cannot access resources outside your company", nil)
		return
	}

	neighbors := toNeighborResponses(output.Neighbors)
	writeJSON(w, http.StatusOK, neighborsCollectionResponse{
		Data: neighbors,
		Meta: neighborsMetaResponse{
			SourceProductID:   output.SourceProductID,
			SourceProductName: output.SourceProductName,
			Count:             len(neighbors),
			Limit:             output.Limit,
		},
	})
}

func (h *Handler) RecordNeighborFeedback(w http.ResponseWriter, r *http.Request) {
	if err := requireJSONContentType(r); err != nil {
		writeProblem(w, r, http.StatusUnsupportedMediaType, "Unsupported Media Type", err.Error(), nil)
		return
	}

	sourceProductID, err := parseIDParam(r)
	if err != nil {
		writeProblem(w, r, http.StatusBadRequest, "Invalid Identifier", "id must be a valid integer", nil)
		return
	}

	suggestedProductID, err := parseNeighborIDParam(r)
	if err != nil {
		writeProblem(w, r, http.StatusBadRequest, "Invalid Identifier", "neighbor_id must be a valid integer", nil)
		return
	}

	authenticatedUser, ok := authhttp.AuthenticatedUserFromContext(r.Context())
	if !ok {
		writeProblem(w, r, http.StatusUnauthorized, "Unauthorized", "missing or invalid bearer token", nil)
		return
	}

	sourceProduct, err := h.useCase.GetByID(r.Context(), sourceProductID)
	if err != nil {
		writeApplicationError(w, r, err)
		return
	}
	if err := authhttp.EnsureCompanyAccess(r.Context(), sourceProduct.CompanyID); err != nil {
		writeProblem(w, r, http.StatusForbidden, "Forbidden", "you cannot access resources outside your company", nil)
		return
	}

	var request neighborFeedbackRequest
	if err := decodeJSONBody(r, &request); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "Invalid Request Body", err.Error(), nil)
		return
	}

	output, err := h.useCase.RecordNeighborFeedback(
		r.Context(),
		toNeighborFeedbackInput(sourceProductID, suggestedProductID, authenticatedUser.ID, request),
	)
	if err != nil {
		writeApplicationError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, neighborFeedbackResourceResponse{Data: toNeighborFeedbackResponse(output)})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	if err := requireJSONContentType(r); err != nil {
		writeProblem(w, r, http.StatusUnsupportedMediaType, "Unsupported Media Type", err.Error(), nil)
		return
	}

	id, err := parseIDParam(r)
	if err != nil {
		writeProblem(w, r, http.StatusBadRequest, "Invalid Identifier", "id must be a valid integer", nil)
		return
	}

	var request updateProductRequest
	if err := decodeJSONBody(r, &request); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "Invalid Request Body", err.Error(), nil)
		return
	}
	if err := authhttp.EnsureCompanyAccess(r.Context(), request.CompanyID); err != nil {
		writeProblem(w, r, http.StatusForbidden, "Forbidden", "you cannot access resources outside your company", nil)
		return
	}

	output, err := h.useCase.Update(r.Context(), toUpdateInput(id, request))
	if err != nil {
		writeApplicationError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, resourceResponse{Data: toProductResponse(output)})
}

func filterProductsByCompany(outputs []productapp.Output, companyID int64) []productapp.Output {
	filtered := make([]productapp.Output, 0, len(outputs))
	for _, output := range outputs {
		if output.CompanyID == companyID {
			filtered = append(filtered, output)
		}
	}
	return filtered
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeProblem(w, r, http.StatusBadRequest, "Invalid Identifier", "id must be a valid integer", nil)
		return
	}

	if err := h.useCase.Delete(r.Context(), id); err != nil {
		writeApplicationError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseIDParam(r *http.Request) (int64, error) {
	return strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
}

func parseNeighborIDParam(r *http.Request) (int64, error) {
	return strconv.ParseInt(mux.Vars(r)["neighbor_id"], 10, 64)
}
