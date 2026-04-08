package salehttp

import (
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterSalesRoutes(r *mux.Router, handler *Handler) {
	if handler == nil {
		return
	}

	r.HandleFunc("/sales", handler.Create).Methods(http.MethodPost)
	r.HandleFunc("/sales", handler.List).Methods(http.MethodGet)
	r.HandleFunc("/sales/{id:[0-9]+}", handler.GetByID).Methods(http.MethodGet)
}
