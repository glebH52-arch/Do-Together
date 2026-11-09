package repository

import (
	"context"
	"do-together/internal/domain"
	"errors"
)

var ErrCommentNotFound = errors.New("comment not found or access denied")

type TaskCommentRepository interface {
	Create(ctx context.Context, actorID, projectID, taskID int, body string) (*domain.TaskComment, error)
	List(ctx context.Context, actorID, projectID, taskID, limit, offset int) ([]*domain.TaskComment, error)
	Update(ctx context.Context, actorID, projectID, taskID int, commentID int64, body string) error
	Delete(ctx context.Context, actorID, projectID, taskID int, commentID int64) error
}
