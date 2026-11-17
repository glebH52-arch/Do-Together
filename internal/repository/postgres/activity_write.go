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
	_, err = tx.Exec(ctx, stmtInsertProjectActivity, projectID, actorID, action, objectType, objectID, objectTitle, string(oldJSON), string(newJSON))
	if err != nil {
		return fmt.Errorf("insert project activity: %w", err)
	}
	return nil
}
