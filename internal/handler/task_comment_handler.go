package handler

import (
	"do-together/internal/domain"
	"do-together/internal/middleware"
	"do-together/internal/service"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

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
	ID         int64      `json:"id"`
	TaskID     int        `json:"task_id"`
	AuthorID   int        `json:"author_id"`
	AuthorName string     `json:"author_name"`
	Body       string     `json:"body"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
}

func commentToResponse(comment *domain.TaskComment) commentResponse {
	return commentResponse{ID: comment.ID, TaskID: comment.TaskID, AuthorID: comment.AuthorID,
		AuthorName: comment.AuthorName, Body: comment.Body, CreatedAt: comment.CreatedAt, UpdatedAt: comment.UpdatedAt}
}

func commentIDs(r *http.Request) (actorID, projectID, taskID int, ok bool) {
	actorID, ok = middleware.UserIDFromContext(r.Context())
	if !ok || actorID <= 0 {
		return 0, 0, 0, false
	}
	projectID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || projectID <= 0 {
		return 0, 0, 0, false
	}
	taskID, err = strconv.Atoi(mux.Vars(r)["taskID"])
	return actorID, projectID, taskID, err == nil && taskID > 0
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
	commentID, err := strconv.ParseInt(mux.Vars(r)["commentID"], 10, 64)
	if !ok || err != nil || commentID <= 0 {
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
	commentID, err := strconv.ParseInt(mux.Vars(r)["commentID"], 10, 64)
	if !ok || err != nil || commentID <= 0 {
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
