package transferhttp

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	authhttp "github.com/IanStuardo-Dev/backend-crud/internal/adapters/http/auth"
	transferapp "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer"
	"github.com/gorilla/mux"
)

type Handler struct {
	useCase transferapp.UseCase
}

func NewHandler(useCase transferapp.UseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if err := requireJSONContentType(r); err != nil {
		writeProblem(w, r, http.StatusUnsupportedMediaType, "Unsupported Media Type", err.Error(), nil)
		return
	}

	var request createTransferRequest
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

	w.Header().Set("Location", fmt.Sprintf("/inventory/transfers/%d", output.ID))
	writeJSON(w, http.StatusCreated, resourceResponse{Data: toTransferResponse(output)})
}

func (h *Handler) Approve(w http.ResponseWriter, r *http.Request) {
	h.transition(w, r, h.useCase.Approve)
}

func (h *Handler) Dispatch(w http.ResponseWriter, r *http.Request) {
	h.transition(w, r, h.useCase.Dispatch)
}

func (h *Handler) Receive(w http.ResponseWriter, r *http.Request) {
	h.transition(w, r, h.useCase.Receive)
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	h.transition(w, r, h.useCase.Cancel)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	outputs, err := h.useCase.List(r.Context())
	if err != nil {
		writeApplicationError(w, r, err)
		return
	}
	if authenticatedUser, ok := authhttp.AuthenticatedUserFromContext(r.Context()); ok && !authenticatedUser.IsSuperAdmin() {
		outputs = filterTransfersByCompany(outputs, *authenticatedUser.CompanyID)
	}

	responses := toTransferResponses(outputs)
	writeJSON(w, http.StatusOK, collectionResponse{
		Data: responses,
		Meta: metaResponse{Count: len(responses)},
	})
}

func (h *Handler) ListByBranch(w http.ResponseWriter, r *http.Request) {
	branchID, err := strconv.ParseInt(mux.Vars(r)["branch_id"], 10, 64)
	if err != nil {
		writeProblem(w, r, http.StatusBadRequest, "Invalid Identifier", "branch_id must be a valid integer", nil)
		return
	}

	outputs, err := h.useCase.ListByBranch(r.Context(), branchID)
	if err != nil {
		writeApplicationError(w, r, err)
		return
	}
	if authenticatedUser, ok := authhttp.AuthenticatedUserFromContext(r.Context()); ok && !authenticatedUser.IsSuperAdmin() {
		outputs = filterTransfersByCompany(outputs, *authenticatedUser.CompanyID)
	}

	responses := toTransferResponses(outputs)
	writeJSON(w, http.StatusOK, collectionResponse{
		Data: responses,
		Meta: metaResponse{Count: len(responses)},
	})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
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

	writeJSON(w, http.StatusOK, resourceResponse{Data: toTransferResponse(output)})
}

func filterTransfersByCompany(outputs []transferapp.Output, companyID int64) []transferapp.Output {
	filtered := make([]transferapp.Output, 0, len(outputs))
	for _, output := range outputs {
		if output.CompanyID == companyID {
			filtered = append(filtered, output)
		}
	}
	return filtered
}

func (h *Handler) transition(w http.ResponseWriter, r *http.Request, action func(context.Context, transferapp.TransitionInput) (transferapp.Output, error)) {
	authenticatedUser, ok := authhttp.AuthenticatedUserFromContext(r.Context())
	if !ok {
		writeProblem(w, r, http.StatusUnauthorized, "Unauthorized", "missing or invalid bearer token", nil)
		return
	}

	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		writeProblem(w, r, http.StatusBadRequest, "Invalid Identifier", "id must be a valid integer", nil)
		return
	}

	output, err := action(r.Context(), transferapp.TransitionInput{
		ID:          id,
		ActorUserID: authenticatedUser.ID,
	})
	if err != nil {
		writeApplicationError(w, r, err)
		return
	}
	if err := authhttp.EnsureCompanyAccess(r.Context(), output.CompanyID); err != nil {
		writeProblem(w, r, http.StatusForbidden, "Forbidden", "you cannot access resources outside your company", nil)
		return
	}

	writeJSON(w, http.StatusOK, resourceResponse{Data: toTransferResponse(output)})
}
