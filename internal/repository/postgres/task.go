package postgres

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.TaskRepository = (*PostgresTaskRepository)(nil)

type PostgresTaskRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) *PostgresTaskRepository {
	return &PostgresTaskRepository{
		pool: pool,
	}

}

func (t *PostgresTaskRepository) Create(ctx context.Context, task *domain.Task) (*domain.Task, error) {

	if task == nil {
		return nil, repository.ErrNilTask
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if task.AssigneeID != nil {
		var member bool
		if err := t.pool.QueryRow(ctx, sqlTaskMemberExists, task.ProjectID, *task.AssigneeID).Scan(&member); err != nil {
			return nil, err
		}
		if !member {
			return nil, repository.ErrAssigneeNotMember
		}
	}

	err := t.pool.QueryRow(ctx, stmtCreateTask, task.ProjectID, task.Title, task.Description, task.CreatedBy, task.Status, task.CreatedAt, task.DueAt,
		task.Type, task.Priority, task.AssigneeID).Scan(&task.ID, &task.UUID, &task.Key, &task.Number, &task.Position, &task.CreatorName, &task.AssigneeName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if task.AssigneeID != nil {
				return nil, repository.ErrAssigneeNotMember
			}
			return nil, repository.ErrProjectNotFound
		}
		return nil, fmt.Errorf("create task: %w", err)
	}
	return task, err
}

func (t *PostgresTaskRepository) List(ctx context.Context, actorID, projectID int) ([]*domain.Task, error) {
	if actorID <= 0 {
		return nil, repository.ErrUserNotFound
	}
	if projectID <= 0 {
		return nil, repository.ErrProjectNotFound
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var member bool
	if err := t.pool.QueryRow(ctx, sqlTaskMemberExists, projectID, actorID).Scan(&member); err != nil {
		return nil, fmt.Errorf("check task list access: %w", err)
	}
	if !member {
		return nil, repository.ErrProjectNotFound
	}
	rows, err := t.pool.Query(ctx, stmtGetListTask, projectID, actorID)
	if err != nil {
		return nil, fmt.Errorf("query list task: %w", err)
	}
	defer rows.Close()

	tasks := make([]*domain.Task, 0)
	for rows.Next() {
		var task domain.Task

		err := rows.Scan(&task.ID, &task.ProjectID, &task.Title, &task.Description, &task.CreatedBy, &task.Status, &task.CreatedAt, &task.UpdatedAt, &task.DueAt,
			&task.UUID, &task.Key, &task.Type, &task.Priority, &task.AssigneeID, &task.AssigneeName, &task.CreatorName, &task.Position, &task.CompletedAt)
		if err != nil {
			return nil, fmt.Errorf("rows scan: %w", err)
		}
		tasks = append(tasks, &task)

	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}
	return tasks, nil
}
func (t *PostgresTaskRepository) Update(ctx context.Context, before, task *domain.Task, actorID, projectID int) error {
	if actorID <= 0 {
		return repository.ErrUserNotFound
	}
	if projectID <= 0 {
		return repository.ErrProjectNotFound
	}
	if task == nil || before == nil {
		return repository.ErrNilTask
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin task update: %w", err)
	}
	defer tx.Rollback(ctx)
	if task.AssigneeID != nil {
		var member bool
		if err := tx.QueryRow(ctx, sqlTaskMemberExists, projectID, *task.AssigneeID).Scan(&member); err != nil {
			return err
		}
		if !member {
			return repository.ErrAssigneeNotMember
		}
	}
	var updatedID int
	err = tx.QueryRow(ctx, stmtUpdateTask, task.Title, task.Description, task.Status, task.UpdatedAt,
		task.ID, projectID, actorID, before.Status, before.UpdatedAt, task.CompletedAt,
		task.Type, task.Priority, task.AssigneeID, task.DueAt).Scan(&updatedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return repository.ErrTaskConflict
	}
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	type historyChange struct {
		action    string
		old, next map[string]any
	}
	changes := make([]historyChange, 0, 5)
	oldValues := make(map[string]any)
	newValues := make(map[string]any)
	if before.Title != task.Title {
		oldValues["title"], newValues["title"] = before.Title, task.Title
	}
	if before.Description != task.Description {
		oldValues["description"], newValues["description"] = before.Description, task.Description
	}
	if before.Type != task.Type {
		oldValues["type"], newValues["type"] = before.Type, task.Type
	}
	if len(oldValues) > 0 {
		changes = append(changes, historyChange{"edited", oldValues, newValues})
	}
	if before.Status != task.Status {
		changes = append(changes, historyChange{"status_changed", map[string]any{"status": before.Status}, map[string]any{"status": task.Status}})
	}
	if before.Priority != task.Priority {
		changes = append(changes, historyChange{"priority_changed", map[string]any{"priority": before.Priority}, map[string]any{"priority": task.Priority}})
	}
	if !reflect.DeepEqual(before.AssigneeID, task.AssigneeID) {
		changes = append(changes, historyChange{"assignee_changed", map[string]any{"assignee_id": before.AssigneeID}, map[string]any{"assignee_id": task.AssigneeID}})
	}
	if !reflect.DeepEqual(before.DueAt, task.DueAt) {
		changes = append(changes, historyChange{"due_at_changed", map[string]any{"due_at": before.DueAt}, map[string]any{"due_at": task.DueAt}})
	}
	for _, change := range changes {
		oldJSON, err := json.Marshal(change.old)
		if err != nil {
			return err
		}
		newJSON, err := json.Marshal(change.next)
		if err != nil {
			return err
		}
		var historyID int64
		if err := tx.QueryRow(ctx, sqlTaskHistoryInsert, updatedID, actorID, change.action, string(oldJSON), string(newJSON)).Scan(&historyID); err != nil {
			return fmt.Errorf("record task history: %w", err)
		}
		var kind string
		var first, second *int
		switch change.action {
		case "assignee_changed":
			kind, first = "assignment_changed", task.AssigneeID
		case "status_changed":
			kind, first, second = "status_changed", &task.CreatedBy, task.AssigneeID
		case "due_at_changed":
			kind, first, second = "due_at_changed", &task.CreatedBy, task.AssigneeID
		}
		if kind != "" {
			if _, err := tx.Exec(ctx, sqlInsertTaskNotifications, projectID, task.ID, historyID, kind, actorID, first, second); err != nil {
				return fmt.Errorf("record task notification: %w", err)
			}
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit task update: %w", err)
	}
	return nil
}
func (t *PostgresTaskRepository) Remove(ctx context.Context, taskID, actorID, projectID int) error {
	if actorID <= 0 {
		return repository.ErrUserNotFound
	}
	if projectID <= 0 {
		return repository.ErrProjectNotFound
	}
	if taskID <= 0 {
		return repository.ErrTaskNotFound
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	commandTag, err := t.pool.Exec(ctx, stmtRemoveTask, taskID, projectID, actorID)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return repository.ErrTaskNotFound
	}
	return nil
}

func (t *PostgresTaskRepository) GetByID(ctx context.Context, actorID, projectID, taskID int) (*domain.Task, error) {
	if actorID <= 0 {
		return nil, repository.ErrUserNotFound
	}
	if projectID <= 0 {
		return nil, repository.ErrProjectNotFound
	}
	if taskID <= 0 {
		return nil, repository.ErrTaskNotFound
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	task := &domain.Task{}

	err := t.pool.QueryRow(ctx, stmtGetTaskByID, taskID, projectID, actorID).
		Scan(
			&task.ID,
			&task.ProjectID,
			&task.Title,
			&task.Description,
			&task.CreatedBy,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.DueAt,
			&task.UUID,
			&task.Key,
			&task.Type,
			&task.Priority,
			&task.AssigneeID,
			&task.AssigneeName,
			&task.CreatorName,
			&task.Position,
			&task.CompletedAt,
		)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrProjectNotFound
		}
		return nil, fmt.Errorf("get task by id: %w", err)
	}
	return task, nil
}
