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

const listProjectActivitySQL = `
WITH access AS (
    SELECT EXISTS (
        SELECT 1 FROM project_members
        WHERE project_id = $1 AND user_id = $2
    ) AS allowed
), page AS (
    SELECT h.id, h.project_id, h.actor_id,
	       h.actor_name, h.action,
           h.object_type, h.object_id, h.object_title,
	       (h.action IN ('task.deleted', 'comment.deleted')
	        OR (h.task_id IS NULL AND h.object_type = 'task')) AS object_deleted,
           h.old_values, h.new_values, h.occurred_at
    FROM task_history AS h
    WHERE h.project_id = $1
      AND ($3::timestamptz IS NULL OR (h.occurred_at, h.id) < ($3, $4::uuid))
    ORDER BY h.occurred_at DESC, h.id DESC
    LIMIT $5
)
SELECT access.allowed, page.id, page.project_id, page.actor_id,
       page.actor_name, page.action, page.object_type, page.object_id,
       page.object_title, page.object_deleted, page.old_values,
       page.new_values, page.occurred_at
FROM access LEFT JOIN page ON access.allowed`

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
	rows, err := r.pool.Query(ctx, listProjectActivitySQL, projectID, actorID, cursorTime, cursorID, limit)
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
