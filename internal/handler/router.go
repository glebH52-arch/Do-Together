package handler

import (
	"do-together/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(p *ProjectHandler, u *UserHandler, a *AuthHandler, au *middleware.AuthMiddleware) http.Handler {
	router := mux.NewRouter()
	router.Path("/projects").Methods(http.MethodPost).Handler(au.Authenticate(http.HandlerFunc(p.CreateProject)))
	router.Path("/projects/{id}").Methods(http.MethodGet).HandlerFunc(p.GetProject)
	router.Path("/projects").Methods(http.MethodGet).HandlerFunc(p.ListProjects)
	router.Path("/projects/{id}").Methods(http.MethodPatch).HandlerFunc(p.UpdateProject)
	router.Path("/users").Methods(http.MethodPost).HandlerFunc(u.CreateUser)
	router.Path("/auth/login").Methods(http.MethodPost).HandlerFunc(a.Login)
	router.Path("/users/me").Methods(http.MethodGet).Handler(au.Authenticate(http.HandlerFunc(u.GetMe)))
	return router
}
