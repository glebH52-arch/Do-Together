package repository

import (
	"context"
	"do-together/internal/domain"
	"errors"
	"time"
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
	List(ctx context.Context, actorID, projectID int) ([]*domain.Task, error)
	Update(ctx context.Context, before, after *domain.Task, actorID, projectID int) error
	Remove(ctx context.Context, taskID, actorID, projectID int) error
	GetByID(ctx context.Context, actorID, projectID, taskID int) (*domain.Task, error)
	Move(ctx context.Context, actorID, projectID, taskID int, expectedStatus domain.TaskStatus, expectedPosition int64, beforeTaskID *int) error
	Search(ctx context.Context, actorID, projectID int, options TaskSearchOptions) ([]*domain.Task, error)
}

type TaskSearchOptions struct {
	Query      string
	Status     string
	AssigneeID *int
	Priority   string
	Type       string
	DueBefore  *time.Time
	DueAfter   *time.Time
	Sort       string
	Direction  string
	Limit      int
	Offset     int
}
