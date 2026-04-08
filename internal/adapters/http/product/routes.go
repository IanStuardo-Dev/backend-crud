package producthttp

import (
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterReadRoutes(r *mux.Router, handler *Handler) {
	if handler == nil {
		return
	}

	r.HandleFunc("/products", handler.List).Methods(http.MethodGet)
	r.HandleFunc("/products/{id:[0-9]+}", handler.GetByID).Methods(http.MethodGet)
	r.HandleFunc("/products/{id:[0-9]+}/neighbors", handler.FindNeighbors).Methods(http.MethodGet)
}

func RegisterWriteRoutes(r *mux.Router, handler *Handler) {
	if handler == nil {
		return
	}

	r.HandleFunc("/products", handler.Create).Methods(http.MethodPost)
	r.HandleFunc("/products/{id:[0-9]+}", handler.Update).Methods(http.MethodPut)
	r.HandleFunc("/products/{id:[0-9]+}", handler.Delete).Methods(http.MethodDelete)
}
