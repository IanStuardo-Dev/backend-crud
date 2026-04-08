package userhttp

import (
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterAdminRoutes(r *mux.Router, handler *Handler) {
	if handler == nil {
		return
	}

	r.HandleFunc("/users", handler.Create).Methods(http.MethodPost)
	r.HandleFunc("/users", handler.List).Methods(http.MethodGet)
	r.HandleFunc("/users/{id:[0-9]+}", handler.GetByID).Methods(http.MethodGet)
	r.HandleFunc("/users/{id:[0-9]+}", handler.Update).Methods(http.MethodPut)
	r.HandleFunc("/users/{id:[0-9]+}", handler.Delete).Methods(http.MethodDelete)
}
