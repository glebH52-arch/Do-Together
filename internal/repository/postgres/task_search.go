package postgres

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"fmt"
)

var taskSortColumns = map[string]string{
	"created_at": "t.created_at",
	"due_at":     "t.due_at",
	"title":      "t.title",
	"priority":   "CASE t.priority WHEN 'low' THEN 0 WHEN 'medium' THEN 1 WHEN 'high' THEN 2 ELSE 3 END",
}

func (t *PostgresTaskRepository) Search(ctx context.Context, actorID, projectID int, options repository.TaskSearchOptions) ([]*domain.Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	column, ok := taskSortColumns[options.Sort]
	if !ok || (options.Direction != "asc" && options.Direction != "desc") || options.Limit <= 0 || options.Limit > 100 || options.Offset < 0 {
		return nil, fmt.Errorf("invalid task search options")
	}
	var member bool
	if err := t.pool.QueryRow(ctx, sqlTaskMemberExists, projectID, actorID).Scan(&member); err != nil {
		return nil, fmt.Errorf("check task search access: %w", err)
	}
	if !member {
		return nil, repository.ErrProjectNotFound
	}
	query := sqlSearchTasksBase + column + ` ` + options.Direction + ` NULLS LAST, t.id ASC LIMIT $10 OFFSET $11`
	rows, err := t.pool.Query(ctx, query, projectID, actorID, options.Query, options.Status, options.AssigneeID,
		options.Priority, options.Type, options.DueBefore, options.DueAfter, options.Limit, options.Offset)
	if err != nil {
		return nil, fmt.Errorf("search tasks: %w", err)
	}
	defer rows.Close()
	tasks := make([]*domain.Task, 0)
	for rows.Next() {
		task := &domain.Task{}
		if err := rows.Scan(&task.ID, &task.ProjectID, &task.Title, &task.Description, &task.CreatedBy, &task.Status,
			&task.CreatedAt, &task.UpdatedAt, &task.DueAt, &task.UUID, &task.Key, &task.Type, &task.Priority,
			&task.AssigneeID, &task.AssigneeName, &task.CreatorName, &task.Position, &task.CompletedAt); err != nil {
			return nil, fmt.Errorf("scan task search: %w", err)
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}
