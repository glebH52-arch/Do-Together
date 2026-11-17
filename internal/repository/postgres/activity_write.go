package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"uuid"

	"github.com/jackc/pgx/v5"
)

func insertProjectActivity(
	ctx context.Context,
	tx pgx.Tx,
	projectID, actorID, objectID uuid.UUID,
	action, objectType, objectTitle string,
	oldValues, newValues any,
) error {
	oldJSON, err := json.Marshal(oldValues)
	if err != nil {
		return fmt.Errorf("marshal activity old values: %w", err)
	}
	newJSON, err := json.Marshal(newValues)
	if err != nil {
		return fmt.Errorf("marshal activity new values: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO task_history (
		project_id, task_id, actor_id, actor_name, action,
		object_type, object_id, object_title, old_values, new_values
	)
	SELECT $1, NULL, actor.id, actor.display_name, $3, $4, $5, $6, $7::jsonb, $8::jsonb
	FROM users AS actor
	WHERE actor.id = $2`, projectID, actorID, action, objectType, objectID, objectTitle, string(oldJSON), string(newJSON))
	if err != nil {
		return fmt.Errorf("insert project activity: %w", err)
	}
	return nil
}
