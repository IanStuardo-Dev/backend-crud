package authhttp

import (
	"net/http"

	authapp "github.com/example/crud/internal/application/auth"
)

type Handler struct {
	useCase authapp.UseCase
}

func NewHandler(useCase authapp.UseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if err := requireJSONContentType(r); err != nil {
		writeProblem(w, r, http.StatusUnsupportedMediaType, "Unsupported Media Type", err.Error(), nil)
		return
	}

	var request loginRequest
	if err := decodeJSONBody(r, &request); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "Invalid Request Body", err.Error(), nil)
		return
	}

	output, err := h.useCase.Login(r.Context(), toLoginInput(request))
	if err != nil {
		writeApplicationError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, resourceResponse{Data: toLoginResponse(output)})
}
