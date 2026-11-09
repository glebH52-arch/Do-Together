package postgres

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type boardItem struct {
	id       int
	position int64
}

func (t *PostgresTaskRepository) Move(ctx context.Context, actorID, projectID, taskID int, expectedStatus domain.TaskStatus, expectedPosition int64, beforeTaskID *int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin task move: %w", err)
	}
	defer tx.Rollback(ctx)
	var lockedProject int
	err = tx.QueryRow(ctx, sqlLockProjectMember, projectID, actorID).Scan(&lockedProject)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.ErrProjectNotFound
		}
		return fmt.Errorf("lock task project: %w", err)
	}
	var status domain.TaskStatus
	var oldPosition int64
	err = tx.QueryRow(ctx, sqlLockTaskPosition, taskID, projectID).Scan(&status, &oldPosition)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.ErrTaskNotFound
		}
		return fmt.Errorf("lock task: %w", err)
	}
	if status != expectedStatus || oldPosition != expectedPosition {
		return repository.ErrTaskConflict
	}
	rows, err := tx.Query(ctx, sqlListTaskPositions, projectID, status)
	if err != nil {
		return fmt.Errorf("read task positions: %w", err)
	}
	items := make([]boardItem, 0)
	for rows.Next() {
		var item boardItem
		if err := rows.Scan(&item.id, &item.position); err != nil {
			rows.Close()
			return err
		}
		items = append(items, item)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	remaining := make([]boardItem, 0, len(items)-1)
	for _, item := range items {
		if item.id != taskID {
			remaining = append(remaining, item)
		}
	}
	insertAt := len(remaining)
	if beforeTaskID != nil && *beforeTaskID != taskID {
		insertAt = -1
		for index, item := range remaining {
			if item.id == *beforeTaskID {
				insertAt = index
				break
			}
		}
		if insertAt < 0 {
			return repository.ErrTaskNotFound
		}
	}
	if beforeTaskID != nil && *beforeTaskID == taskID {
		return nil
	}
	currentIndex := -1
	for index, item := range items {
		if item.id == taskID {
			currentIndex = index
			break
		}
	}
	if currentIndex == insertAt {
		return nil
	}
	var newPosition int64
	rebalance := false
	switch {
	case len(remaining) == 0:
		return nil
	case insertAt == 0:
		if remaining[0].position > 1 {
			newPosition = remaining[0].position / 2
		} else {
			rebalance = true
		}
	case insertAt == len(remaining):
		last := remaining[len(remaining)-1].position
		if last < 1<<62 {
			newPosition = last + 1024
		} else {
			rebalance = true
		}
	default:
		low, high := remaining[insertAt-1].position, remaining[insertAt].position
		if high-low > 1 {
			newPosition = low + (high-low)/2
		} else {
			rebalance = true
		}
	}
	if rebalance {
		ordered := make([]boardItem, 0, len(items))
		ordered = append(ordered, remaining[:insertAt]...)
		ordered = append(ordered, boardItem{id: taskID})
		ordered = append(ordered, remaining[insertAt:]...)
		for index, item := range ordered {
			position := int64(index+1) * 1024
			if item.id == taskID {
				newPosition = position
			}
			if _, err := tx.Exec(ctx, sqlSetTaskPosition, position, item.id); err != nil {
				return fmt.Errorf("rebalance task positions: %w", err)
			}
		}
	} else {
		if _, err := tx.Exec(ctx, sqlSetTaskPosition, newPosition, taskID); err != nil {
			return fmt.Errorf("move task: %w", err)
		}
	}
	if _, err := tx.Exec(ctx, sqlTouchTask, taskID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, sqlTaskMoveHistory,
		taskID, actorID, oldPosition, newPosition); err != nil {
		return fmt.Errorf("record task move: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit task move: %w", err)
	}
	return nil
}
