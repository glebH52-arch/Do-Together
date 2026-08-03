package repository

import (
	"context"
	"do-together/internal/domain"
	"errors"
)

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrNilTask      = errors.New("nil task")
)

type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task) (*domain.Task, error)
	List(ctx context.Context, actorID, projectID int) ([]*domain.Task, error)
	Update(ctx context.Context, task *domain.Task, actorID, projectID int) error
	Remove(ctx context.Context, taskID, actorID, projectID int) error
	GetByID(ctx context.Context, actorID, projectID, taskID int) (*domain.Task, error)
}
