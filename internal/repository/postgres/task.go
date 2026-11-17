package postgres

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"uuid"

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
	newValues, err := json.Marshal(map[string]any{
		"title":       task.Title,
		"status":      task.Status,
		"type":        task.Type,
		"priority":    task.Priority,
		"assignee_id": task.AssigneeID,
		"due_at":      task.DueAt,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal created task history: %w", err)
	}
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin task create: %w", err)
	}
	defer tx.Rollback(ctx)
	if task.AssigneeID != nil {
		var member bool
		err = tx.QueryRow(ctx, stmtLockTaskAssignee, task.ProjectID, *task.AssigneeID).Scan(&member)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrAssigneeNotMember
		}
		if err != nil {
			return nil, fmt.Errorf("lock task assignee: %w", err)
		}
	}
	var taskID uuid.UUID
	err = tx.QueryRow(ctx, stmtCreateTask, task.ProjectID, task.Title, task.Description, task.CreatedBy, task.Status,
		task.CreatedAt, task.DueAt, task.Type, task.Priority, task.AssigneeID).Scan(&taskID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrProjectNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("insert task: %w", err)
	}
	var historyID uuid.UUID
	err = tx.QueryRow(ctx, stmtInsertTaskCreatedHistory,
		taskID, task.ProjectID, task.CreatedBy, task.Title, string(newValues)).Scan(&historyID)
	if err != nil {
		return nil, fmt.Errorf("record task creation: %w", err)
	}
	if task.AssigneeID != nil {
		if _, err := tx.Exec(ctx, stmtInsertTaskNotifications, task.ProjectID, taskID, historyID,
			string(domain.NotificationKindAssignmentChanged), task.CreatedBy, task.AssigneeID, nil); err != nil {
			return nil, fmt.Errorf("record task assignment notification: %w", err)
		}
	}
	var creatorName string
	var assigneeName *string
	err = tx.QueryRow(ctx, stmtGetTaskParticipants, task.CreatedBy, task.AssigneeID).Scan(
		&creatorName, &task.CreatorAvatarFilename, &task.CreatorAvatarUpdatedAt,
		&assigneeName, &task.AssigneeAvatarFilename, &task.AssigneeAvatarUpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get task participants: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit task create: %w", err)
	}
	task.ID = taskID
	task.CreatorName = creatorName
	task.AssigneeName = assigneeName
	return task, nil
}

func (t *PostgresTaskRepository) List(ctx context.Context, actorID, projectID uuid.UUID) ([]*domain.Task, error) {
	if actorID == uuid.Nil() {
		return nil, repository.ErrUserNotFound
	}
	if projectID == uuid.Nil() {
		return nil, repository.ErrProjectNotFound
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var member bool
	if err := t.pool.QueryRow(ctx, stmtTaskMemberExists, projectID, actorID).Scan(&member); err != nil {
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
			&task.Type, &task.Priority, &task.AssigneeID, &task.AssigneeName,
			&task.AssigneeAvatarFilename, &task.AssigneeAvatarUpdatedAt,
			&task.CreatorName, &task.CreatorAvatarFilename, &task.CreatorAvatarUpdatedAt, &task.CompletedAt)
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
func (t *PostgresTaskRepository) Update(ctx context.Context, before, task *domain.Task, actorID, projectID uuid.UUID) error {
	if actorID == uuid.Nil() {
		return repository.ErrUserNotFound
	}
	if projectID == uuid.Nil() {
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

		err := tx.QueryRow(ctx, stmtLockTaskAssignee, projectID, *task.AssigneeID).Scan(&member)

		if errors.Is(err, pgx.ErrNoRows) {
			return repository.ErrAssigneeNotMember
		}
		if err != nil {
			return fmt.Errorf("lock task assignee: %w", err)
		}
	}
	var updatedID uuid.UUID
	err = tx.QueryRow(ctx, stmtUpdateTask, task.Title, task.Description, task.Status, task.UpdatedAt,
		task.ID, projectID, actorID, before.Status, before.UpdatedAt, task.CompletedAt,
		task.Type, task.Priority, task.AssigneeID, task.DueAt).Scan(&updatedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return repository.ErrTaskConflict
	}
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	changes := make(map[string]any)
	addChange := func(field string, oldValue, newValue any) {
		changes[field] = map[string]any{"old": oldValue, "new": newValue}
	}
	if before.Title != task.Title {
		addChange("title", before.Title, task.Title)
	}
	if before.Description != task.Description {
		changes["description"] = map[string]any{"changed": true}
	}
	if before.Type != task.Type {
		addChange("type", before.Type, task.Type)
	}
	if before.Status != task.Status {
		addChange("status", before.Status, task.Status)
	}
	if before.Priority != task.Priority {
		addChange("priority", before.Priority, task.Priority)
	}
	if !reflect.DeepEqual(before.AssigneeID, task.AssigneeID) {
		addChange("assignee_id", before.AssigneeID, task.AssigneeID)
	}
	if !reflect.DeepEqual(before.DueAt, task.DueAt) {
		addChange("due_at", before.DueAt, task.DueAt)
	}
	if len(changes) > 0 {
		newJSON, err := json.Marshal(map[string]any{"changes": changes})
		if err != nil {
			return err
		}
		var historyID uuid.UUID
		if err := tx.QueryRow(ctx, stmtTaskHistoryInsert,
			updatedID, projectID, actorID, "task.updated", task.Title, nil, string(newJSON)).Scan(&historyID); err != nil {
			return fmt.Errorf("record task history: %w", err)
		}
		notifications := []struct {
			changed       bool
			kind          domain.NotificationKind
			first, second *uuid.UUID
		}{
			{!reflect.DeepEqual(before.AssigneeID, task.AssigneeID), domain.NotificationKindAssignmentChanged, before.AssigneeID, task.AssigneeID},
			{before.Status != task.Status, domain.NotificationKindStatusChanged, &task.CreatedBy, task.AssigneeID},
			{!reflect.DeepEqual(before.DueAt, task.DueAt), domain.NotificationKindDueAtChanged, &task.CreatedBy, task.AssigneeID},
		}
		for _, notification := range notifications {
			if !notification.changed {
				continue
			}
			if _, err := tx.Exec(ctx, stmtInsertTaskNotifications, projectID, task.ID, historyID,
				string(notification.kind), actorID, notification.first, notification.second); err != nil {
				return fmt.Errorf("record task notification: %w", err)
			}
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit task update: %w", err)
	}
	return nil
}
func (t *PostgresTaskRepository) Remove(ctx context.Context, taskID, actorID, projectID uuid.UUID) error {
	if actorID == uuid.Nil() {
		return repository.ErrUserNotFound
	}
	if projectID == uuid.Nil() {
		return repository.ErrProjectNotFound
	}
	if taskID == uuid.Nil() {
		return repository.ErrTaskNotFound
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin task delete: %w", err)
	}
	defer tx.Rollback(ctx)
	var title string
	if err := tx.QueryRow(ctx, stmtLockTaskForDelete, taskID, actorID, projectID).Scan(&title); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.ErrTaskNotFound
		}
		return fmt.Errorf("lock task for delete: %w", err)
	}
	deletedValues, err := json.Marshal(map[string]any{"deleted": true})
	if err != nil {
		return err
	}
	var historyID uuid.UUID
	if err := tx.QueryRow(ctx, stmtTaskHistoryInsert,
		taskID, projectID, actorID, "task.deleted", title, nil, string(deletedValues)).Scan(&historyID); err != nil {
		return fmt.Errorf("record task deletion: %w", err)
	}
	commandTag, err := tx.Exec(ctx, stmtRemoveTask, taskID, projectID, actorID)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return repository.ErrTaskNotFound
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit task delete: %w", err)
	}
	return nil
}

func (t *PostgresTaskRepository) GetByID(ctx context.Context, actorID, projectID, taskID uuid.UUID) (*domain.Task, error) {
	if actorID == uuid.Nil() {
		return nil, repository.ErrUserNotFound
	}
	if projectID == uuid.Nil() {
		return nil, repository.ErrProjectNotFound
	}
	if taskID == uuid.Nil() {
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
			&task.Type,
			&task.Priority,
			&task.AssigneeID,
			&task.AssigneeName,
			&task.AssigneeAvatarFilename,
			&task.AssigneeAvatarUpdatedAt,
			&task.CreatorName,
			&task.CreatorAvatarFilename,
			&task.CreatorAvatarUpdatedAt,
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

func (t *PostgresTaskRepository) History(ctx context.Context, actorID, projectID, taskID uuid.UUID, limit, offset int) ([]*domain.TaskHistory, error) {
	rows, err := t.pool.Query(ctx, `SELECT h.id,h.task_id,h.actor_id,u.username,h.action,h.old_values,h.new_values,h.occurred_at FROM task_history h JOIN users u ON u.id=h.actor_id JOIN tasks t ON t.id=h.task_id JOIN project_members pm ON pm.project_id=t.project_id AND pm.user_id=$1 WHERE t.project_id=$2 AND t.id=$3 ORDER BY h.occurred_at DESC,h.id DESC LIMIT $4 OFFSET $5`, actorID, projectID, taskID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]*domain.TaskHistory, 0)
	for rows.Next() {
		var h domain.TaskHistory
		if err := rows.Scan(&h.ID, &h.TaskID, &h.ActorID, &h.ActorName, &h.Action, &h.OldValues, &h.NewValues, &h.OccurredAt); err != nil {
			return nil, err
		}
		items = append(items, &h)
	}
	return items, rows.Err()
}
