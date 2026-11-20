package postgres

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"encoding/json"
	"fmt"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresProjectActivityRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresProjectActivityRepository(pool *pgxpool.Pool) *PostgresProjectActivityRepository {
	return &PostgresProjectActivityRepository{pool: pool}
}

func (r *PostgresProjectActivityRepository) List(ctx context.Context, actorID, projectID uuid.UUID, cursor *repository.ActivityCursor, limit int) ([]*domain.ProjectActivity, error) {
	var cursorTime *time.Time
	var cursorID *uuid.UUID
	if cursor != nil {
		cursorTime = &cursor.OccurredAt
		cursorID = &cursor.ID
	}
	rows, err := r.pool.Query(ctx, stmtListProjectActivity, projectID, actorID, cursorTime, cursorID, limit)
	if err != nil {
		return nil, fmt.Errorf("list project activity: %w", err)
	}
	defer rows.Close()
	items := make([]*domain.ProjectActivity, 0, limit)
	allowed := false
	for rows.Next() {
		var rowAllowed bool
		var id, rowProjectID, rowActorID, objectID *uuid.UUID
		var actorName, eventType, objectType, objectTitle *string
		var objectDeleted *bool
		var oldValues, newValues *[]byte
		var occurredAt *time.Time
		if err := rows.Scan(&rowAllowed, &id, &rowProjectID, &rowActorID, &actorName, &eventType, &objectType, &objectID, &objectTitle, &objectDeleted, &oldValues, &newValues, &occurredAt); err != nil {
			return nil, fmt.Errorf("scan project activity: %w", err)
		}
		allowed = rowAllowed
		if id == nil {
			continue
		}
		item := &domain.ProjectActivity{
			ID: *id, ProjectID: *rowProjectID, ActorID: *rowActorID,
			ActorName: *actorName, Type: domain.PublicActivityType(*eventType, *objectType),
			ObjectType: *objectType, ObjectID: *objectID, ObjectTitle: *objectTitle,
			ObjectDeleted: *objectDeleted, OccurredAt: *occurredAt,
		}
		if oldValues != nil {
			item.OldValues = json.RawMessage(*oldValues)
		}
		if newValues != nil {
			item.NewValues = json.RawMessage(*newValues)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate project activity: %w", err)
	}
	if !allowed {
		return nil, repository.ErrProjectNotFound
	}
	return items, nil
}
