package handler

import (
	"bytes"
	"do-together/internal/domain"
	"do-together/internal/middleware"
	"do-together/internal/service"
	"encoding/json"
	"net/http"
	"uuid"

	"github.com/gorilla/mux"
)

type InviteHandler struct {
	inviteService *service.InviteService
}

func NewInviteHandler(inviteService *service.InviteService) *InviteHandler {
	return &InviteHandler{
		inviteService: inviteService,
	}
}

func (h *InviteHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == uuid.Nil() {
		status := http.StatusUnauthorized
		http.Error(w, http.StatusText(status), status)
		return
	}
	idText := mux.Vars(r)["id"]
	id, err := uuid.Parse(idText)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if id == uuid.Nil() {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	request := createInviteRequest{}
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if (request.InviteeID != uuid.Nil()) == (request.Email != "") {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	var invite *domain.Invite
	if request.Email != "" {
		invite, err = h.inviteService.CreateByEmail(r.Context(), userID, id, request.Email, domain.ProjectMemberRole(request.Role))
	} else {
		invite, err = h.inviteService.Create(r.Context(), userID, request.InviteeID, id, domain.ProjectMemberRole(request.Role))
	}
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	response := inviteToResponse(invite)
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

func (h *InviteHandler) ListInvites(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == uuid.Nil() {
		status := http.StatusUnauthorized
		http.Error(w, http.StatusText(status), status)
		return
	}
	invites, err := h.inviteService.List(r.Context(), userID)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	response := make([]inviteResponse, 0, len(invites))
	for _, invite := range invites {
		response = append(response, inviteToResponse(invite))
	}
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
func (h *InviteHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == uuid.Nil() {
		status := http.StatusUnauthorized
		http.Error(w, http.StatusText(status), status)
		return
	}
	idText := mux.Vars(r)["id"]
	id, err := uuid.Parse(idText)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if id == uuid.Nil() {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	err = h.inviteService.Accept(r.Context(), userID, id)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *InviteHandler) DeclineInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == uuid.Nil() {
		status := http.StatusUnauthorized
		http.Error(w, http.StatusText(status), status)
		return
	}
	idText := mux.Vars(r)["id"]
	id, err := uuid.Parse(idText)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if id == uuid.Nil() {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	err = h.inviteService.Decline(r.Context(), userID, id)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
