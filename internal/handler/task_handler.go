package handler

import (
	"bytes"
	"do-together/internal/domain"
	"do-together/internal/middleware"
	"do-together/internal/repository"
	"do-together/internal/service"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

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
	if !ok || userID <= 0 {
		status := http.StatusUnauthorized
		http.Error(w, http.StatusText(status), status)
		return
	}
	idText := mux.Vars(r)["id"]
	projectID, err := strconv.Atoi(idText)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if projectID <= 0 {
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
	task, err := t.taskService.CreateTaskWithOptions(r.Context(), userID, projectID, request.Title, request.Description, request.DueAt, request.Type, request.Priority, request.AssigneeID)
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
	if !ok || userID <= 0 {
		status := http.StatusUnauthorized
		http.Error(w, http.StatusText(status), status)
		return
	}
	idText := mux.Vars(r)["id"]
	projectID, err := strconv.Atoi(idText)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if projectID <= 0 {
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
	if !ok || userID <= 0 {
		status := http.StatusUnauthorized
		http.Error(w, http.StatusText(status), status)
		return
	}
	idText := mux.Vars(r)["id"]
	projectID, err := strconv.Atoi(idText)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if projectID <= 0 {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	idText = mux.Vars(r)["taskID"]
	taskID, err := strconv.Atoi(idText)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if taskID <= 0 {
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
	assignee, err := parseOptionalInt(request.AssigneeID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	dueAt, err := parseOptionalTime(request.DueAt)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	err = t.taskService.UpdateTaskChanges(r.Context(), userID, projectID, taskID, domain.TaskChanges{
		Title: request.Title, Description: request.Description, Status: status,
		Type: request.Type, Priority: request.Priority, AssigneeID: assignee, DueAt: dueAt,
	})
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (t *TaskHandler) RemoveTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID <= 0 {
		status := http.StatusUnauthorized
		http.Error(w, http.StatusText(status), status)
		return
	}
	idText := mux.Vars(r)["id"]
	projectID, err := strconv.Atoi(idText)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if projectID <= 0 {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	idText = mux.Vars(r)["taskID"]
	taskID, err := strconv.Atoi(idText)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if taskID <= 0 {
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
	if !ok || userID <= 0 {
		status := http.StatusUnauthorized
		http.Error(w, http.StatusText(status), status)
		return
	}
	idText := mux.Vars(r)["id"]
	projectID, err := strconv.Atoi(idText)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if projectID <= 0 {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	idText = mux.Vars(r)["taskID"]
	taskID, err := strconv.Atoi(idText)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	if taskID <= 0 {
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
	if !ok || userID <= 0 {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	projectID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || projectID <= 0 {
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

func (t *TaskHandler) MoveTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID <= 0 {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	projectID, projectErr := strconv.Atoi(mux.Vars(r)["id"])
	taskID, taskErr := strconv.Atoi(mux.Vars(r)["taskID"])
	if projectErr != nil || taskErr != nil || projectID <= 0 || taskID <= 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	var request taskMoveRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.ExpectedPosition <= 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if request.BeforeTaskID != nil && *request.BeforeTaskID <= 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if err := t.taskService.MoveTask(r.Context(), userID, projectID, taskID, request.ExpectedStatus, request.ExpectedPosition, request.BeforeTaskID); err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (t *TaskHandler) SearchTasks(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || actorID <= 0 {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	projectID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || projectID <= 0 {
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
			var id int
			id, err = strconv.Atoi(value)
			options.AssigneeID = &id
		}
	}
	if err == nil {
		if value := query.Get("due_before"); value != "" {
			var due time.Time
			due, err = time.Parse(time.RFC3339, value)
			options.DueBefore = &due
		}
	}
	if err == nil {
		if value := query.Get("due_after"); value != "" {
			var due time.Time
			due, err = time.Parse(time.RFC3339, value)
			options.DueAfter = &due
		}
	}
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	tasks, err := t.taskService.SearchTasks(r.Context(), actorID, projectID, options)
	if err != nil {
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
