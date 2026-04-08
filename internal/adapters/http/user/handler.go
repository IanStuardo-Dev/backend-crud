package userhttp

import (
	"fmt"
	"net/http"
	"strconv"

	authhttp "github.com/example/crud/internal/adapters/http/auth"
	userapp "github.com/example/crud/internal/application/user"
	domainuser "github.com/example/crud/internal/domain/user"
	"github.com/gorilla/mux"
)

type Handler struct {
	useCase userapp.UseCase
}

func NewHandler(useCase userapp.UseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if err := requireJSONContentType(r); err != nil {
		writeProblem(w, r, http.StatusUnsupportedMediaType, "Unsupported Media Type", err.Error(), nil)
		return
	}

	var request createUserRequest
	if err := decodeJSONBody(r, &request); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "Invalid Request Body", err.Error(), nil)
		return
	}

	if authenticatedUser, ok := authhttp.AuthenticatedUserFromContext(r.Context()); ok {
		if request.CompanyID == nil || !authenticatedUser.CanAccessCompany(*request.CompanyID) {
			writeProblem(w, r, http.StatusForbidden, "Forbidden", "you cannot manage users outside your company", nil)
			return
		}
		if !authenticatedUser.IsSuperAdmin() && request.Role == domainuser.RoleSuperAdmin {
			writeProblem(w, r, http.StatusForbidden, "Forbidden", "you cannot assign the super_admin role", nil)
			return
		}
	}

	output, err := h.useCase.Create(r.Context(), toCreateInput(request))
	if err != nil {
		writeApplicationError(w, r, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/users/%d", output.ID))
	writeJSON(w, http.StatusCreated, resourceResponse{Data: toUserResponse(output)})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	outputs, err := h.useCase.List(r.Context())
	if err != nil {
		writeApplicationError(w, r, err)
		return
	}

	if authenticatedUser, ok := authhttp.AuthenticatedUserFromContext(r.Context()); ok && !authenticatedUser.IsSuperAdmin() {
		outputs = filterUsersByCompany(outputs, *authenticatedUser.CompanyID)
	}

	responses := toUserResponses(outputs)
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

	if authenticatedUser, ok := authhttp.AuthenticatedUserFromContext(r.Context()); ok && !authenticatedUser.IsSuperAdmin() {
		if output.CompanyID == nil || !authenticatedUser.CanAccessCompany(*output.CompanyID) {
			writeProblem(w, r, http.StatusForbidden, "Forbidden", "you cannot access users outside your company", nil)
			return
		}
	}

	writeJSON(w, http.StatusOK, resourceResponse{Data: toUserResponse(output)})
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

	var request updateUserRequest
	if err := decodeJSONBody(r, &request); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "Invalid Request Body", err.Error(), nil)
		return
	}

	if authenticatedUser, ok := authhttp.AuthenticatedUserFromContext(r.Context()); ok {
		if request.CompanyID == nil || !authenticatedUser.CanAccessCompany(*request.CompanyID) {
			writeProblem(w, r, http.StatusForbidden, "Forbidden", "you cannot manage users outside your company", nil)
			return
		}
		if !authenticatedUser.IsSuperAdmin() && request.Role == domainuser.RoleSuperAdmin {
			writeProblem(w, r, http.StatusForbidden, "Forbidden", "you cannot assign the super_admin role", nil)
			return
		}
	}

	output, err := h.useCase.Update(r.Context(), toUpdateInput(id, request))
	if err != nil {
		writeApplicationError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, resourceResponse{Data: toUserResponse(output)})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeProblem(w, r, http.StatusBadRequest, "Invalid Identifier", "id must be a valid integer", nil)
		return
	}

	if authenticatedUser, ok := authhttp.AuthenticatedUserFromContext(r.Context()); ok && !authenticatedUser.IsSuperAdmin() {
		output, err := h.useCase.GetByID(r.Context(), id)
		if err != nil {
			writeApplicationError(w, r, err)
			return
		}
		if output.CompanyID == nil || !authenticatedUser.CanAccessCompany(*output.CompanyID) {
			writeProblem(w, r, http.StatusForbidden, "Forbidden", "you cannot manage users outside your company", nil)
			return
		}
	}

	if err := h.useCase.Delete(r.Context(), id); err != nil {
		writeApplicationError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func filterUsersByCompany(outputs []userapp.Output, companyID int64) []userapp.Output {
	filtered := make([]userapp.Output, 0, len(outputs))
	for _, output := range outputs {
		if output.CompanyID != nil && *output.CompanyID == companyID {
			filtered = append(filtered, output)
		}
	}
	return filtered
}

func parseIDParam(r *http.Request) (int64, error) {
	return strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
}
