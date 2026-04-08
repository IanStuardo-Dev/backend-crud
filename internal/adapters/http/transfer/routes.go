package transferhttp

import (
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterTransferRoutes(r *mux.Router, handler *Handler) {
	if handler == nil {
		return
	}

	r.HandleFunc("/inventory/transfers", handler.Create).Methods(http.MethodPost)
	r.HandleFunc("/inventory/transfers", handler.List).Methods(http.MethodGet)
	r.HandleFunc("/inventory/transfers/branches/{branch_id:[0-9]+}", handler.ListByBranch).Methods(http.MethodGet)
	r.HandleFunc("/inventory/transfers/{id:[0-9]+}", handler.GetByID).Methods(http.MethodGet)
	r.HandleFunc("/inventory/transfers/{id:[0-9]+}/approve", handler.Approve).Methods(http.MethodPost)
	r.HandleFunc("/inventory/transfers/{id:[0-9]+}/dispatch", handler.Dispatch).Methods(http.MethodPost)
	r.HandleFunc("/inventory/transfers/{id:[0-9]+}/receive", handler.Receive).Methods(http.MethodPost)
	r.HandleFunc("/inventory/transfers/{id:[0-9]+}/cancel", handler.Cancel).Methods(http.MethodPost)
}
