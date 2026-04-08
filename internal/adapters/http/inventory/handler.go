package inventoryhttp

import (
	"net/http"
	"strconv"

	authhttp "github.com/example/crud/internal/adapters/http/auth"
	inventoryapp "github.com/example/crud/internal/application/inventory"
)

type Handler struct {
	useCase inventoryapp.UseCase
}

func NewHandler(useCase inventoryapp.UseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) ListByBranch(w http.ResponseWriter, r *http.Request) {
	request, err := parseListByBranchRequest(r)
	if err != nil {
		writeProblem(w, r, http.StatusBadRequest, "Invalid Query Parameters", err.Error(), nil)
		return
	}
	if err := authhttp.EnsureCompanyAccess(r.Context(), request.CompanyID); err != nil {
		writeProblem(w, r, http.StatusForbidden, "Forbidden", "you cannot access resources outside your company", nil)
		return
	}

	outputs, err := h.useCase.ListByBranch(r.Context(), toListByBranchInput(request))
	if err != nil {
		writeApplicationError(w, r, err)
		return
	}

	responses := toItemResponses(outputs)
	writeJSON(w, http.StatusOK, collectionResponse{
		Data: responses,
		Meta: metaResponse{Count: len(responses)},
	})
}

func (h *Handler) SuggestSources(w http.ResponseWriter, r *http.Request) {
	request, err := parseSuggestSourcesRequest(r)
	if err != nil {
		writeProblem(w, r, http.StatusBadRequest, "Invalid Query Parameters", err.Error(), nil)
		return
	}
	if err := authhttp.EnsureCompanyAccess(r.Context(), request.CompanyID); err != nil {
		writeProblem(w, r, http.StatusForbidden, "Forbidden", "you cannot access resources outside your company", nil)
		return
	}

	outputs, err := h.useCase.SuggestSources(r.Context(), toSuggestSourcesInput(request))
	if err != nil {
		writeApplicationError(w, r, err)
		return
	}

	responses := toSourceCandidateResponses(outputs)
	writeJSON(w, http.StatusOK, collectionResponse{
		Data: responses,
		Meta: metaResponse{Count: len(responses)},
	})
}

func parseListByBranchRequest(r *http.Request) (listByBranchRequest, error) {
	companyID, err := parseInt64Query(r, "company_id")
	if err != nil {
		return listByBranchRequest{}, err
	}
	branchID, err := parseInt64Query(r, "branch_id")
	if err != nil {
		return listByBranchRequest{}, err
	}

	return listByBranchRequest{
		CompanyID: companyID,
		BranchID:  branchID,
	}, nil
}

func parseSuggestSourcesRequest(r *http.Request) (suggestSourcesRequest, error) {
	companyID, err := parseInt64Query(r, "company_id")
	if err != nil {
		return suggestSourcesRequest{}, err
	}
	destinationBranchID, err := parseInt64Query(r, "destination_branch_id")
	if err != nil {
		return suggestSourcesRequest{}, err
	}
	productID, err := parseInt64Query(r, "product_id")
	if err != nil {
		return suggestSourcesRequest{}, err
	}
	quantity, err := parseIntQuery(r, "quantity")
	if err != nil {
		return suggestSourcesRequest{}, err
	}

	return suggestSourcesRequest{
		CompanyID:           companyID,
		DestinationBranchID: destinationBranchID,
		ProductID:           productID,
		Quantity:            quantity,
	}, nil
}

func parseInt64Query(r *http.Request, key string) (int64, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return 0, &queryError{Field: key, Message: key + " is required"}
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, &queryError{Field: key, Message: key + " must be a valid integer"}
	}

	return parsed, nil
}

func parseIntQuery(r *http.Request, key string) (int, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return 0, &queryError{Field: key, Message: key + " is required"}
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, &queryError{Field: key, Message: key + " must be a valid integer"}
	}

	return parsed, nil
}

type queryError struct {
	Field   string
	Message string
}

func (e *queryError) Error() string {
	return e.Message
}
