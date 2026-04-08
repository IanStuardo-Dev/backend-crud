package authhttp

import (
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterPublicRoutes(r *mux.Router, handler *Handler) {
	if handler == nil {
		return
	}

	r.HandleFunc("/auth/login", handler.Login).Methods(http.MethodPost)
}
