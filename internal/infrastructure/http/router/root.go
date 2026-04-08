package router

import (
	"net/http"

	"github.com/gorilla/mux"
)

func configureRootHandlers(r *mux.Router) {
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		writeProblem(w, req, http.StatusNotFound, "Resource Not Found", "endpoint not found")
	})
	r.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		writeProblem(w, req, http.StatusMethodNotAllowed, "Method Not Allowed", "method not allowed for this endpoint")
	})
}
