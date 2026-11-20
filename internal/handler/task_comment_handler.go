package handler

import (
	"do-together/internal/domain"
	"do-together/internal/middleware"
	"do-together/internal/service"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"uuid"

	"github.com/gorilla/mux"
)

type TaskCommentHandler struct {
	service *service.TaskCommentService
}

func NewTaskCommentHandler(service *service.TaskCommentService) *TaskCommentHandler {
	return &TaskCommentHandler{service: service}
}

type commentRequest struct {
	Body string `json:"body"`
}

type commentResponse struct {
	ID         uuid.UUID  `json:"id"`
	TaskID     uuid.UUID  `json:"task_id"`
	AuthorID   uuid.UUID  `json:"author_id"`
	AuthorName string     `json:"author_name"`
	Body       string     `json:"body"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
}

func commentToResponse(comment *domain.TaskComment) commentResponse {
	return commentResponse{ID: comment.ID, TaskID: comment.TaskID, AuthorID: comment.AuthorID,
		AuthorName: comment.AuthorName, Body: comment.Body, CreatedAt: comment.CreatedAt, UpdatedAt: comment.UpdatedAt}
}

func commentIDs(r *http.Request) (actorID, projectID, taskID uuid.UUID, ok bool) {
	actorID, ok = middleware.UserIDFromContext(r.Context())
	if !ok || actorID == uuid.Nil() {
		return uuid.Nil(), uuid.Nil(), uuid.Nil(), false
	}
	projectID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil || projectID == uuid.Nil() {
		return uuid.Nil(), uuid.Nil(), uuid.Nil(), false
	}
	taskID, err = uuid.Parse(mux.Vars(r)["taskID"])
	return actorID, projectID, taskID, err == nil && taskID != uuid.Nil()
}

func (h *TaskCommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	actorID, projectID, taskID, ok := commentIDs(r)
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	var request commentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	comment, err := h.service.Create(r.Context(), actorID, projectID, taskID, request.Body)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(commentToResponse(comment))
}

func (h *TaskCommentHandler) List(w http.ResponseWriter, r *http.Request) {
	actorID, projectID, taskID, ok := commentIDs(r)
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	limit, offset := 20, 0
	var err error
	if value := r.URL.Query().Get("limit"); value != "" {
		limit, err = strconv.Atoi(value)
	}
	if err == nil {
		if value := r.URL.Query().Get("offset"); value != "" {
			offset, err = strconv.Atoi(value)
		}
	}
	if err != nil || limit <= 0 || limit > 100 || offset < 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	comments, err := h.service.List(r.Context(), actorID, projectID, taskID, limit, offset)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	response := make([]commentResponse, 0, len(comments))
	for _, comment := range comments {
		response = append(response, commentToResponse(comment))
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h *TaskCommentHandler) Update(w http.ResponseWriter, r *http.Request) {
	actorID, projectID, taskID, ok := commentIDs(r)
	commentID, err := uuid.Parse(mux.Vars(r)["commentID"])
	if !ok || err != nil || commentID == uuid.Nil() {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	var request commentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if err := h.service.Update(r.Context(), actorID, projectID, taskID, commentID, request.Body); err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskCommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	actorID, projectID, taskID, ok := commentIDs(r)
	commentID, err := uuid.Parse(mux.Vars(r)["commentID"])
	if !ok || err != nil || commentID == uuid.Nil() {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if err := h.service.Delete(r.Context(), actorID, projectID, taskID, commentID); err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
