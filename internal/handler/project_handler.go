package handler

import (
	"bytes"
	"do-together/internal/middleware"
	"do-together/internal/service"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type ProjectHandler struct {
	projectService *service.ProjectService
}

func NewProjectHandler(p *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectService: p,
	}
}

func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID <= 0 {
		status := http.StatusUnauthorized
		http.Error(w, http.StatusText(status), status)
		return
	}
	request := projectRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	project, err := h.projectService.CreateProject(r.Context(), userID, request.Title, request.Goal)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	response := projectToResponse(project)
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(response)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(buf.Bytes())
	if err != nil {
		return
	}
}
func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID <= 0 {
		status := http.StatusUnauthorized
		http.Error(w, http.StatusText(status), status)
		return
	}
	idText := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idText)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if id <= 0 {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	project, err := h.projectService.GetProject(r.Context(), userID, id)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	response := projectToResponse(project)
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(response)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(buf.Bytes())
	if err != nil {
		return
	}
}
func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID <= 0 {
		status := http.StatusUnauthorized
		http.Error(w, http.StatusText(status), status)
		return
	}
	projects, err := h.projectService.ListProjects(r.Context(), userID)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	projectsResponse := make([]projectResponse, 0, len(projects))
	for _, project := range projects {
		projectsResponse = append(projectsResponse, projectToResponse(project))
	}
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(projectsResponse)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(buf.Bytes())
	if err != nil {
		return
	}
}
func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID <= 0 {
		status := http.StatusUnauthorized
		http.Error(w, http.StatusText(status), status)
		return
	}
	idText := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idText)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if id <= 0 {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	request := updateProjectRequest{}
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if request.Title == nil && request.Goal == nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	project, err := h.projectService.UpdateProject(r.Context(), userID, id, request.Title, request.Goal)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	response := projectToResponse(project)
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(response)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(buf.Bytes())
	if err != nil {
		return
	}
}
