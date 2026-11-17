package postgres

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"fmt"
	"uuid"
)

var taskSortColumns = map[string]struct{}{
	"created_at": {},
	"due_at":     {},
	"title":      {},
	"priority":   {},
}

func (t *PostgresTaskRepository) Search(ctx context.Context, actorID, projectID uuid.UUID, options repository.TaskSearchOptions) ([]*domain.Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	_, ok := taskSortColumns[options.Sort]
	if !ok || (options.Direction != "asc" && options.Direction != "desc") || options.Limit <= 0 || options.Limit > 100 || options.Offset < 0 {
		return nil, fmt.Errorf("invalid task search options")
	}
	var member bool
	if err := t.pool.QueryRow(ctx, stmtTaskMemberExists, projectID, actorID).Scan(&member); err != nil {
		return nil, fmt.Errorf("check task search access: %w", err)
	}
	if !member {
		return nil, repository.ErrProjectNotFound
	}
	rows, err := t.pool.Query(ctx, stmtSearchTasks, projectID, actorID, options.Query, options.Status, options.AssigneeID,
		options.Priority, options.Type, options.DueBefore, options.DueAfter, options.Sort, options.Direction, options.Limit, options.Offset)
	if err != nil {
		return nil, fmt.Errorf("search tasks: %w", err)
	}
	defer rows.Close()
	tasks := make([]*domain.Task, 0)
	for rows.Next() {
		task := &domain.Task{}
		if err := rows.Scan(&task.ID, &task.ProjectID, &task.Title, &task.Description, &task.CreatedBy, &task.Status,
			&task.CreatedAt, &task.UpdatedAt, &task.DueAt, &task.Type, &task.Priority,
			&task.AssigneeID, &task.AssigneeName, &task.AssigneeAvatarFilename, &task.AssigneeAvatarUpdatedAt,
			&task.CreatorName, &task.CreatorAvatarFilename, &task.CreatorAvatarUpdatedAt, &task.CompletedAt); err != nil {
			return nil, fmt.Errorf("scan task search: %w", err)
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}
