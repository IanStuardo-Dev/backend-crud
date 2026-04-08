package inventoryhttp

import (
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterInventoryRoutes(r *mux.Router, handler *Handler) {
	if handler == nil {
		return
	}

	r.HandleFunc("/inventory/branch-items", handler.ListByBranch).Methods(http.MethodGet)
	r.HandleFunc("/inventory/source-candidates", handler.SuggestSources).Methods(http.MethodGet)
}
