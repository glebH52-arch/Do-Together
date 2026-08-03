package postgres

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
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

	err := t.pool.QueryRow(ctx, stmtCreateTask, task.ProjectID, task.Title, task.Description, task.CreatedBy, task.Status, task.CreatedAt, task.DueAt).Scan(&task.ID)
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
	rows, err := t.pool.Query(ctx, stmtGetListTask, projectID, actorID)
	if err != nil {
		return nil, fmt.Errorf("query list task: %w", err)
	}
	defer rows.Close()

	tasks := make([]*domain.Task, 0)
	for rows.Next() {
		var task domain.Task

		err := rows.Scan(&task.ID, &task.ProjectID, &task.Title, &task.Description, &task.CreatedBy, &task.Status, &task.CreatedAt, &task.UpdatedAt, &task.DueAt)
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
func (t *PostgresTaskRepository) Update(ctx context.Context, task *domain.Task, actorID, projectID int) error {
	if actorID <= 0 {
		return repository.ErrUserNotFound
	}
	if projectID <= 0 {
		return repository.ErrProjectNotFound
	}
	if task == nil {
		return repository.ErrNilTask
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	commandTag, err := t.pool.Exec(ctx, stmtUpdateTask, task.Title, task.Description, task.Status, task.UpdatedAt, task.ID, projectID, actorID)

	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return repository.ErrTaskNotFound
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
		)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrProjectNotFound
		}
		return nil, fmt.Errorf("get task by id: %w", err)
	}
	return task, nil
}
