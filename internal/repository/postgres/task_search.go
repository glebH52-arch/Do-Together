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
	if err := t.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM project_members WHERE project_id = $1 AND user_id = $2)`, projectID, actorID).Scan(&member); err != nil {
		return nil, fmt.Errorf("check task search access: %w", err)
	}
	if !member {
		return nil, repository.ErrProjectNotFound
	}
	query := `SELECT t.id, t.project_id, t.title, t.description, t.created_by, t.status,
		t.created_at, t.updated_at, t.due_at, t.uuid::text, t.task_key, t.task_type, t.priority,
		t.assignee_id, assignee.username, creator.username, t.position, t.completed_at
		FROM tasks AS t JOIN users AS creator ON creator.id = t.created_by
		LEFT JOIN users AS assignee ON assignee.id = t.assignee_id
		WHERE t.project_id = $1
		AND EXISTS (SELECT 1 FROM project_members AS pm WHERE pm.project_id = t.project_id AND pm.user_id = $2)
		AND ($3::text = '' OR t.title ILIKE '%' || $3 || '%' OR t.description ILIKE '%' || $3 || '%')
		AND ($4::text = '' OR t.status = $4)
		AND ($5::int IS NULL OR t.assignee_id = $5)
		AND ($6::text = '' OR t.priority = $6)
		AND ($7::text = '' OR t.task_type = $7)
		AND ($8::timestamptz IS NULL OR t.due_at <= $8)
		AND ($9::timestamptz IS NULL OR t.due_at >= $9)
		ORDER BY ` + column + ` ` + options.Direction + ` NULLS LAST, t.id ASC LIMIT $10 OFFSET $11`
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
