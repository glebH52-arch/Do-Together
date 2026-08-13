package handler

import (
	"do-together/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(p *ProjectHandler, u *UserHandler, a *AuthHandler, au *middleware.AuthMiddleware, pm *ProjectMemberHandler, ih *InviteHandler, t *TaskHandler, h *HealthHandler) http.Handler {
	router := mux.NewRouter()
	router.Path("/projects").Methods(http.MethodPost).Handler(au.Authenticate(http.HandlerFunc(p.CreateProject)))
	router.Path("/projects/{id}").Methods(http.MethodGet).Handler(au.Authenticate(http.HandlerFunc(p.GetProject)))
	router.Path("/projects").Methods(http.MethodGet).Handler(au.Authenticate(http.HandlerFunc(p.ListProjects)))
	router.Path("/projects/{id}").Methods(http.MethodPatch).Handler(au.Authenticate(http.HandlerFunc(p.UpdateProject)))
	router.Path("/projects/{id}").Methods(http.MethodDelete).Handler(au.Authenticate(http.HandlerFunc(p.ArchiveProject)))

	router.Path("/projects/{id}/members").Methods(http.MethodGet).Handler(au.Authenticate(http.HandlerFunc(pm.GetMembers)))
	router.Path("/projects/{id}/members/{userID}").Methods(http.MethodPatch).Handler(au.Authenticate(http.HandlerFunc(pm.UpdateMember)))
	router.Path("/projects/{id}/members/{userID}").Methods(http.MethodDelete).Handler(au.Authenticate(http.HandlerFunc(pm.RemoveMember)))

	router.Path("/projects/{id}/invites").Methods(http.MethodPost).Handler(au.Authenticate(http.HandlerFunc(ih.CreateInvite)))
	router.Path("/invites").Methods(http.MethodGet).Handler(au.Authenticate(http.HandlerFunc(ih.ListInvites)))
	router.Path("/invites/{id}/accept").Methods(http.MethodPost).Handler(au.Authenticate(http.HandlerFunc(ih.AcceptInvite)))
	router.Path("/invites/{id}/decline").Methods(http.MethodPost).Handler(au.Authenticate(http.HandlerFunc(ih.DeclineInvite)))

	router.Path("/projects/{id}/tasks").Methods(http.MethodPost).Handler(au.Authenticate(http.HandlerFunc(t.CreateTask)))
	router.Path("/projects/{id}/tasks").Methods(http.MethodGet).Handler(au.Authenticate(http.HandlerFunc(t.ListTask)))
	router.Path("/projects/{id}/tasks/{taskID}").Methods(http.MethodGet).Handler(au.Authenticate(http.HandlerFunc(t.GetTaskByID)))
	router.Path("/projects/{id}/tasks/{taskID}").Methods(http.MethodPatch).Handler(au.Authenticate(http.HandlerFunc(t.UpdateTask)))
	router.Path("/projects/{id}/tasks/{taskID}").Methods(http.MethodDelete).Handler(au.Authenticate(http.HandlerFunc(t.RemoveTask)))

	router.Path("/users").Methods(http.MethodPost).HandlerFunc(u.CreateUser)
	router.Path("/auth/login").Methods(http.MethodPost).HandlerFunc(a.Login)
	router.Path("/auth/refresh").Methods(http.MethodPost).HandlerFunc(a.Refresh)
	router.Path("/auth/logout").Methods(http.MethodPost).HandlerFunc(a.Logout)
	router.Path("/users/me").Methods(http.MethodGet).Handler(au.Authenticate(http.HandlerFunc(u.GetMe)))
	router.Path("/health").Methods(http.MethodGet).HandlerFunc(h.Health)
	router.Path("/ready").Methods(http.MethodGet).HandlerFunc(h.Ready)
	return router
}
