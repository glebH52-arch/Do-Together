package handler

import (
	"bytes"
	"do-together/internal/domain"
	"do-together/internal/middleware"
	"do-together/internal/repository"
	"do-together/internal/service"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
	"uuid"

	"github.com/gorilla/mux"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler(p *service.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: p,
	}
}

func (t *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == uuid.Nil() {
		status := http.StatusUnauthorized
		http.Error(w, http.StatusText(status), status)
		return
	}
	idText := mux.Vars(r)["id"]
	projectID, err := uuid.Parse(idText)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if projectID == uuid.Nil() {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	request := taskRequest{}
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	params := domain.NewTaskParams{
		CreatedBy:   userID,
		ProjectID:   projectID,
		Title:       request.Title,
		Description: request.Description,
		DueAt:       request.DueAt,
		Type:        request.Type,
		Priority:    request.Priority,
		AssigneeID:  request.AssigneeID,
	}
	task, err := t.taskService.CreateTask(r.Context(), params)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	response := taskToResponse(task)
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
func (t *TaskHandler) ListTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == uuid.Nil() {
		status := http.StatusUnauthorized
		http.Error(w, http.StatusText(status), status)
		return
	}
	idText := mux.Vars(r)["id"]
	projectID, err := uuid.Parse(idText)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if projectID == uuid.Nil() {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	tasks, err := t.taskService.ListTask(r.Context(), userID, projectID)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	tasksResponse := make([]taskResponse, 0, len(tasks))
	for _, task := range tasks {
		tasksResponse = append(tasksResponse, taskToResponse(task))
	}
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(tasksResponse)
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
func (t *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == uuid.Nil() {
		status := http.StatusUnauthorized
		http.Error(w, http.StatusText(status), status)
		return
	}
	idText := mux.Vars(r)["id"]
	projectID, err := uuid.Parse(idText)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if projectID == uuid.Nil() {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	idText = mux.Vars(r)["taskID"]
	taskID, err := uuid.Parse(idText)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if taskID == uuid.Nil() {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	request := taskUpdateRequest{}
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if request.Title == nil &&
		request.Description == nil &&
		request.Status == nil && request.Type == nil && request.Priority == nil &&
		request.AssigneeID == nil && request.DueAt == nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	var status *domain.TaskStatus

	if request.Status != nil {
		value := domain.TaskStatus(*request.Status)
		status = &value
	}
	assignee, err := parseOptionalUUID(request.AssigneeID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	dueAt, err := parseOptionalTime(request.DueAt)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	err = t.taskService.UpdateTask(r.Context(), userID, projectID, taskID, domain.TaskChanges{
		Title: request.Title, Description: request.Description, Status: status,
		Type: request.Type, Priority: request.Priority, AssigneeID: assignee, DueAt: dueAt,
	})
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	updated, err := t.taskService.GetTaskByID(r.Context(), userID, projectID, taskID)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(taskToResponse(updated))
}

func (t *TaskHandler) TaskHistory(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", 401)
		return
	}
	projectID, e1 := uuid.Parse(mux.Vars(r)["id"])
	taskID, e2 := uuid.Parse(mux.Vars(r)["taskID"])
	limit, offset := 50, 0
	var err error
	if v := r.URL.Query().Get("limit"); v != "" {
		limit, err = strconv.Atoi(v)
	}
	if err == nil {
		if v := r.URL.Query().Get("offset"); v != "" {
			offset, err = strconv.Atoi(v)
		}
	}
	if e1 != nil || e2 != nil || err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}
	items, err := t.taskService.TaskHistory(r.Context(), actorID, projectID, taskID, limit, offset)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	type response struct {
		ID        uuid.UUID       `json:"id"`
		TaskID    uuid.UUID       `json:"task_id"`
		ActorID   uuid.UUID       `json:"actor_id"`
		ActorName string          `json:"actor_name"`
		Action    string          `json:"action"`
		Old       json.RawMessage `json:"old_values"`
		New       json.RawMessage `json:"new_values"`
		Created   time.Time       `json:"created_at"`
	}
	out := make([]response, 0, len(items))
	for _, h := range items {
		out = append(out, response{h.ID, h.TaskID, h.ActorID, h.ActorName, h.Action, h.OldValues, h.NewValues, h.OccurredAt})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}
func (t *TaskHandler) RemoveTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == uuid.Nil() {
		status := http.StatusUnauthorized
		http.Error(w, http.StatusText(status), status)
		return
	}
	idText := mux.Vars(r)["id"]
	projectID, err := uuid.Parse(idText)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if projectID == uuid.Nil() {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	idText = mux.Vars(r)["taskID"]
	taskID, err := uuid.Parse(idText)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if taskID == uuid.Nil() {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	err = t.taskService.RemoveTask(r.Context(), taskID, userID, projectID)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (t *TaskHandler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == uuid.Nil() {
		status := http.StatusUnauthorized
		http.Error(w, http.StatusText(status), status)
		return
	}
	idText := mux.Vars(r)["id"]
	projectID, err := uuid.Parse(idText)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if projectID == uuid.Nil() {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	idText = mux.Vars(r)["taskID"]
	taskID, err := uuid.Parse(idText)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if taskID == uuid.Nil() {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	task, err := t.taskService.GetTaskByID(r.Context(), userID, projectID, taskID)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	response := taskToResponse(task)
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

func (t *TaskHandler) GetBoard(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == uuid.Nil() {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	projectID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil || projectID == uuid.Nil() {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	tasks, err := t.taskService.ListTask(r.Context(), userID, projectID)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	board := boardResponse{
		Todo: make([]taskResponse, 0), InProgress: make([]taskResponse, 0),
		Review: make([]taskResponse, 0), Done: make([]taskResponse, 0),
	}
	for _, task := range tasks {
		response := taskToResponse(task)
		switch task.Status {
		case domain.TaskStatusTodo:
			board.Todo = append(board.Todo, response)
		case domain.TaskStatusInProgress:
			board.InProgress = append(board.InProgress, response)
		case domain.TaskStatusReview:
			board.Review = append(board.Review, response)
		case domain.TaskStatusDone:
			board.Done = append(board.Done, response)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(board); err != nil {
		return
	}
}

func (t *TaskHandler) SearchTasks(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || actorID == uuid.Nil() {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	projectID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil || projectID == uuid.Nil() {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	query := r.URL.Query()
	options := repository.TaskSearchOptions{
		Query: query.Get("q"), Status: query.Get("status"), Priority: query.Get("priority"),
		Type: query.Get("type"), Sort: query.Get("sort"), Direction: query.Get("direction"),
		Limit: 20, Offset: 0,
	}
	if value := query.Get("limit"); value != "" {
		options.Limit, err = strconv.Atoi(value)
	}
	if err == nil {
		if value := query.Get("offset"); value != "" {
			options.Offset, err = strconv.Atoi(value)
		}
	}
	if err == nil {
		if value := query.Get("assignee_id"); value != "" {
			var id uuid.UUID
			id, err = uuid.Parse(value)
			options.AssigneeID = &id
		}
	}
	if err == nil {
		if value := query.Get("due_before"); value != "" {
			var due time.Time
			due, err = time.Parse(time.RFC3339, value)
			if err != nil {
				http.Error(w, "due_before must be an RFC3339 timestamp", http.StatusBadRequest)
				return
			}
			options.DueBefore = &due
		}
	}
	if err == nil {
		if value := query.Get("due_after"); value != "" {
			var due time.Time
			due, err = time.Parse(time.RFC3339, value)
			if err != nil {
				http.Error(w, "due_after must be an RFC3339 timestamp", http.StatusBadRequest)
				return
			}
			options.DueAfter = &due
		}
	}
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	tasks, err := t.taskService.SearchTasks(r.Context(), actorID, projectID, options)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidTaskSearch) &&
			options.DueAfter != nil && options.DueBefore != nil &&
			!options.DueAfter.Before(*options.DueBefore) {
			http.Error(w, "due_after must be earlier than due_before", http.StatusBadRequest)
			return
		}
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	response := make([]taskResponse, 0, len(tasks))
	for _, task := range tasks {
		response = append(response, taskToResponse(task))
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
