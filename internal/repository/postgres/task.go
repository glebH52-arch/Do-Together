package postgres

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"encoding/json"
	"errors"
	"fmt"

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

	err := t.pool.QueryRow(ctx, stmtCreateTask, task.ProjectID, task.Title, task.Description, task.CreatedBy, task.Status, task.CreatedAt, task.DueAt).Scan(&task.ID, &task.UUID, &task.Key, &task.Number, &task.Position, &task.CreatorName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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
	if err := t.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM project_members WHERE project_id = $1 AND user_id = $2)`, projectID, actorID).Scan(&member); err != nil {
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
	var updatedID int
	err = tx.QueryRow(ctx, stmtUpdateTask, task.Title, task.Description, task.Status, task.UpdatedAt,
		task.ID, projectID, actorID, before.Status, before.UpdatedAt, task.CompletedAt).Scan(&updatedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return repository.ErrTaskConflict
	}
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	oldValues := make(map[string]any)
	newValues := make(map[string]any)
	if before.Title != task.Title {
		oldValues["title"], newValues["title"] = before.Title, task.Title
	}
	if before.Description != task.Description {
		oldValues["description"], newValues["description"] = before.Description, task.Description
	}
	action := "edited"
	if before.Status != task.Status {
		oldValues["status"], newValues["status"] = before.Status, task.Status
		action = "status_changed"
	}
	oldJSON, err := json.Marshal(oldValues)
	if err != nil {
		return err
	}
	newJSON, err := json.Marshal(newValues)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO task_history(task_id, actor_id, action, old_values, new_values)
		VALUES ($1, $2, $3, $4::jsonb, $5::jsonb)`, updatedID, actorID, action, string(oldJSON), string(newJSON))
	if err != nil {
		return fmt.Errorf("record task history: %w", err)
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
