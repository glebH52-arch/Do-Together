package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(p *ProjectHandler) http.Handler {
	router := mux.NewRouter()
	router.Path("/projects").Methods(http.MethodPost).HandlerFunc(p.CreateProject)
	return router
}
