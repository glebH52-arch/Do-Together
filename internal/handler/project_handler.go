package handler

import (
	"bytes"
	"do-together/internal/service"
	"encoding/json"
	"net/http"
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
	request := createProjectRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	project, err := h.projectService.CreateProject(r.Context(), request.CreatedBy, request.Title, request.Goal)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	response := createProjectResponse{
		ID:        project.ID,
		CreatedBy: project.CreatedBy,
		Title:     project.Title,
		Goal:      project.Goal,
		Status:    string(project.Status),
		CreatedAt: project.CreatedAt,
	}
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
