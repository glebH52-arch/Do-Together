package repository

import (
	"context"
	"do-together/internal/domain"
	"errors"
	"uuid"
)

var ErrCommentNotFound = errors.New("comment not found or access denied")

type TaskCommentRepository interface {
	Create(ctx context.Context, actorID, projectID, taskID uuid.UUID, body string) (*domain.TaskComment, error)
	List(ctx context.Context, actorID, projectID, taskID uuid.UUID, limit, offset int) ([]*domain.TaskComment, error)
	Update(ctx context.Context, actorID, projectID, taskID uuid.UUID, commentID uuid.UUID, body string) error
	Delete(ctx context.Context, actorID, projectID, taskID uuid.UUID, commentID uuid.UUID) error
}
