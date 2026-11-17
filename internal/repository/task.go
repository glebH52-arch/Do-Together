package repository

import (
	"context"
	"do-together/internal/domain"
	"errors"
	"time"
	"uuid"
)

var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrNilTask           = errors.New("nil task")
	ErrTaskConflict      = errors.New("task changed concurrently")
	ErrInvalidTaskSearch = errors.New("invalid task search options")
	ErrAssigneeNotMember = errors.New("assignee is not a project member")
)

type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task) (*domain.Task, error)
	List(ctx context.Context, actorID, projectID uuid.UUID) ([]*domain.Task, error)
	Update(ctx context.Context, before, after *domain.Task, actorID, projectID uuid.UUID) error
	Remove(ctx context.Context, taskID, actorID, projectID uuid.UUID) error
	GetByID(ctx context.Context, actorID, projectID, taskID uuid.UUID) (*domain.Task, error)
	Search(ctx context.Context, actorID, projectID uuid.UUID, options TaskSearchOptions) ([]*domain.Task, error)
	History(ctx context.Context, actorID, projectID, taskID uuid.UUID, limit, offset int) ([]*domain.TaskHistory, error)
}

type TaskSearchOptions struct {
	Query      string
	Status     string
	AssigneeID *uuid.UUID
	Priority   string
	Type       string
	DueBefore  *time.Time
	DueAfter   *time.Time
	Sort       string
	Direction  string
	Limit      int
	Offset     int
}
