package service

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"strings"
	"time"
	"unicode/utf8"
)

type TaskService struct {
	repository repository.TaskRepository
}

func NewTaskService(repository repository.TaskRepository) *TaskService {
	return &TaskService{
		repository: repository,
	}
}

func (s *TaskService) CreateTask(ctx context.Context, actorID, projectID int, title, description string, dueAt *time.Time) (*domain.Task, error) {
	if actorID <= 0 {
		return nil, repository.ErrUserNotFound
	}
	if projectID <= 0 {
		return nil, repository.ErrProjectNotFound
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	task, err := domain.NewTask(actorID, projectID, title, description, dueAt)
	if err != nil {
		return nil, err
	}
	task, err = s.repository.Create(ctx, task)
	if err != nil {
		return nil, err
	}
	return task, nil
}
func (s *TaskService) UpdateTask(ctx context.Context, actorID, projectID, taskID int, title, description *string, status *domain.TaskStatus) error {
	if actorID <= 0 {
		return repository.ErrUserNotFound
	}
	if projectID <= 0 {
		return repository.ErrProjectNotFound
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	task, err := s.repository.GetByID(ctx, actorID, projectID, taskID)
	if err != nil {
		return err
	}
	before := *task
	err = task.Update(title, description, status)
	if err != nil {
		return err
	}
	if task.Title == before.Title && task.Description == before.Description && task.Status == before.Status {
		return nil
	}
	err = s.repository.Update(ctx, &before, task, actorID, projectID)
	if err != nil {
		return err
	}
	return nil
}
func (s *TaskService) RemoveTask(ctx context.Context, taskID, actorID, projectID int) error {
	if actorID <= 0 {
		return repository.ErrUserNotFound
	}
	if projectID <= 0 {
		return repository.ErrProjectNotFound
	}
	if taskID <= 0 {
		return repository.ErrTaskNotFound
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	err := s.repository.Remove(ctx, taskID, actorID, projectID)
	if err != nil {
		return err
	}
	return nil
}
func (s *TaskService) ListTask(ctx context.Context, actorID, projectID int) ([]*domain.Task, error) {
	if actorID <= 0 {
		return nil, repository.ErrUserNotFound
	}
	if projectID <= 0 {
		return nil, repository.ErrProjectNotFound
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	tasks, err := s.repository.List(ctx, actorID, projectID)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *TaskService) GetTaskByID(ctx context.Context, actorID, projectID, taskID int) (*domain.Task, error) {
	if actorID <= 0 {
		return nil, repository.ErrUserNotFound
	}
	if projectID <= 0 {
		return nil, repository.ErrProjectNotFound
	}
	if taskID <= 0 {
		return nil, repository.ErrTaskNotFound
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	task, err := s.repository.GetByID(ctx, actorID, projectID, taskID)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) MoveTask(ctx context.Context, actorID, projectID, taskID int, expectedStatus domain.TaskStatus, expectedPosition int64, beforeTaskID *int) error {
	if actorID <= 0 || projectID <= 0 || taskID <= 0 || expectedPosition <= 0 {
		return repository.ErrTaskNotFound
	}
	if expectedStatus != domain.TaskStatusTodo && expectedStatus != domain.TaskStatusInProgress && expectedStatus != domain.TaskStatusReview && expectedStatus != domain.TaskStatusDone {
		return domain.ErrInvalidTaskStatus
	}
	return s.repository.Move(ctx, actorID, projectID, taskID, expectedStatus, expectedPosition, beforeTaskID)
}

func (s *TaskService) SearchTasks(ctx context.Context, actorID, projectID int, options repository.TaskSearchOptions) ([]*domain.Task, error) {
	if actorID <= 0 || projectID <= 0 || options.Limit <= 0 || options.Limit > 100 || options.Offset < 0 {
		return nil, repository.ErrInvalidTaskSearch
	}
	options.Query = strings.TrimSpace(options.Query)
	if utf8.RuneCountInString(options.Query) > 200 {
		return nil, repository.ErrInvalidTaskSearch
	}
	if options.Status != "" && options.Status != string(domain.TaskStatusTodo) && options.Status != string(domain.TaskStatusInProgress) && options.Status != string(domain.TaskStatusReview) && options.Status != string(domain.TaskStatusDone) {
		return nil, repository.ErrInvalidTaskSearch
	}
	if options.Priority != "" && options.Priority != "low" && options.Priority != "medium" && options.Priority != "high" && options.Priority != "critical" {
		return nil, repository.ErrInvalidTaskSearch
	}
	if options.Type != "" && options.Type != "task" && options.Type != "bug" && options.Type != "feature" {
		return nil, repository.ErrInvalidTaskSearch
	}
	if options.AssigneeID != nil && *options.AssigneeID <= 0 {
		return nil, repository.ErrInvalidTaskSearch
	}
	if options.DueBefore != nil && options.DueAfter != nil && options.DueBefore.Before(*options.DueAfter) {
		return nil, repository.ErrInvalidTaskSearch
	}
	if options.Sort == "" {
		options.Sort = "created_at"
	}
	if options.Direction == "" {
		options.Direction = "desc"
	}
	if options.Sort != "created_at" && options.Sort != "due_at" && options.Sort != "title" && options.Sort != "priority" {
		return nil, repository.ErrInvalidTaskSearch
	}
	if options.Direction != "asc" && options.Direction != "desc" {
		return nil, repository.ErrInvalidTaskSearch
	}
	return s.repository.Search(ctx, actorID, projectID, options)
}
