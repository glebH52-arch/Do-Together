package service

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"time"
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
	err = task.Update(title, description, status)
	if err != nil {
		return err
	}
	err = s.repository.Update(ctx, task, actorID, projectID)
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
