package repository

import (
	"context"
	"do-together/internal/domain"
	"time"
	"uuid"
)

type ActivityCursor struct {
	OccurredAt time.Time
	ID         uuid.UUID
}

type ProjectActivityRepository interface {
	List(ctx context.Context, actorID, projectID uuid.UUID, cursor *ActivityCursor, limit int) ([]*domain.ProjectActivity, error)
}
