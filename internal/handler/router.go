package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(p *ProjectHandler, u *UserHandler) http.Handler {
	router := mux.NewRouter()
	router.Path("/projects").Methods(http.MethodPost).HandlerFunc(p.CreateProject)
	router.Path("/projects/{id}").Methods(http.MethodGet).HandlerFunc(p.GetProject)
	router.Path("/projects").Methods(http.MethodGet).HandlerFunc(p.ListProjects)
	router.Path("/projects/{id}").Methods(http.MethodPatch).HandlerFunc(p.UpdateProject)
	router.Path("/users").Methods(http.MethodPost).HandlerFunc(u.CreateUser)
	return router
}
