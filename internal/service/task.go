package service

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"strings"
	"unicode/utf8"
	"uuid"
)

type TaskService struct {
	repository repository.TaskRepository
}

func NewTaskService(repository repository.TaskRepository) *TaskService {
	return &TaskService{
		repository: repository,
	}
}

func (s *TaskService) CreateTask(ctx context.Context, params domain.NewTaskParams) (*domain.Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	task, err := domain.NewTask(params)
	if err != nil {
		return nil, err
	}
	task, err = s.repository.Create(ctx, task)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, actorID, projectID, taskID uuid.UUID, changes domain.TaskChanges) error {
	if actorID == uuid.Nil() {
		return repository.ErrUserNotFound
	}
	if projectID == uuid.Nil() {
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

	changed, err := task.ApplyChanges(changes)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	err = s.repository.Update(ctx, &before, task, actorID, projectID)
	if err != nil {
		return err
	}
	return nil
}
func (s *TaskService) RemoveTask(ctx context.Context, taskID, actorID, projectID uuid.UUID) error {
	if actorID == uuid.Nil() {
		return repository.ErrUserNotFound
	}
	if projectID == uuid.Nil() {
		return repository.ErrProjectNotFound
	}
	if taskID == uuid.Nil() {
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
func (s *TaskService) ListTask(ctx context.Context, actorID, projectID uuid.UUID) ([]*domain.Task, error) {
	if actorID == uuid.Nil() {
		return nil, repository.ErrUserNotFound
	}
	if projectID == uuid.Nil() {
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

func (s *TaskService) GetTaskByID(ctx context.Context, actorID, projectID, taskID uuid.UUID) (*domain.Task, error) {
	if actorID == uuid.Nil() {
		return nil, repository.ErrUserNotFound
	}
	if projectID == uuid.Nil() {
		return nil, repository.ErrProjectNotFound
	}
	if taskID == uuid.Nil() {
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

func (s *TaskService) SearchTasks(ctx context.Context, actorID, projectID uuid.UUID, options repository.TaskSearchOptions) ([]*domain.Task, error) {
	if actorID == uuid.Nil() || projectID == uuid.Nil() || options.Limit <= 0 || options.Limit > 100 || options.Offset < 0 {
		return nil, repository.ErrInvalidTaskSearch
	}
	options.Query = strings.TrimSpace(options.Query)
	if utf8.RuneCountInString(options.Query) > 200 {
		return nil, repository.ErrInvalidTaskSearch
	}
	if options.Status != "" {
		status := domain.TaskStatus(options.Status)

		if err := domain.ValidateTaskStatus(status); err != nil {
			return nil, repository.ErrInvalidTaskSearch
		}
	}
	if options.Priority != "" {
		if err := domain.ValidateTaskPriority(options.Priority); err != nil {
			return nil, repository.ErrInvalidTaskSearch
		}
	}
	if options.Type != "" {
		if err := domain.ValidateTaskType(options.Type); err != nil {
			return nil, repository.ErrInvalidTaskSearch
		}
	}
	if options.AssigneeID != nil && *options.AssigneeID == uuid.Nil() {
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

func (s *TaskService) TaskHistory(ctx context.Context, actorID, projectID, taskID uuid.UUID, limit, offset int) ([]*domain.TaskHistory, error) {
	if actorID == uuid.Nil() || projectID == uuid.Nil() || taskID == uuid.Nil() || limit < 1 || limit > 100 || offset < 0 {
		return nil, repository.ErrTaskNotFound
	}
	return s.repository.History(ctx, actorID, projectID, taskID, limit, offset)
}
