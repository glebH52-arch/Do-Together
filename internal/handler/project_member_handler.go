package handler

import (
	"bytes"
	"do-together/internal/domain"
	"do-together/internal/middleware"
	"do-together/internal/service"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type ProjectMemberHandler struct {
	projectMemberService *service.ProjectMemberService
}

func NewProjectMemberHandler(p *service.ProjectMemberService) *ProjectMemberHandler {
	return &ProjectMemberHandler{
		projectMemberService: p,
	}
}

func (h *ProjectMemberHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
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
	projectMembers, err := h.projectMemberService.List(r.Context(), userID, id)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	projectsMembersResponse := make([]responseProjectMember, 0, len(projectMembers))
	for _, projectMember := range projectMembers {
		projectsMembersResponse = append(projectsMembersResponse, projectMemberToResponse(projectMember))
	}
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(projectsMembersResponse)
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
func (h *ProjectMemberHandler) AddMember(w http.ResponseWriter, r *http.Request) {
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
	request := requestProjectMember{}
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	err = h.projectMemberService.Add(r.Context(), userID, id, request.UserID, domain.ProjectMemberRole(request.Role))
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
func (h *ProjectMemberHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
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
	idText = mux.Vars(r)["userID"]
	targetID, err := strconv.Atoi(idText)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if targetID <= 0 {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	err = h.projectMemberService.Remove(r.Context(), userID, id, targetID)

	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
func (h *ProjectMemberHandler) UpdateMember(w http.ResponseWriter, r *http.Request) {
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
	idText = mux.Vars(r)["userID"]
	targetID, err := strconv.Atoi(idText)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if targetID <= 0 {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	request := updateProjectMemberRoleRequest{}
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	err = h.projectMemberService.UpdateRole(r.Context(), userID, id, targetID, domain.ProjectMemberRole(request.Role))

	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
