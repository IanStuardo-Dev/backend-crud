package router

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type problemDetail struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
	Path   string `json:"path,omitempty"`
}

func writeProblem(w http.ResponseWriter, req *http.Request, status int, title, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(problemDetail{
		Type:   "https://httpstatuses.com/" + strconv.Itoa(status),
		Title:  title,
		Status: status,
		Detail: detail,
		Path:   req.URL.Path,
	})
}
