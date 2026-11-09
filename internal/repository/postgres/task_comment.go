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

var _ repository.TaskCommentRepository = (*PostgresTaskCommentRepository)(nil)

type PostgresTaskCommentRepository struct {
	pool *pgxpool.Pool
}

func NewTaskCommentRepository(pool *pgxpool.Pool) *PostgresTaskCommentRepository {
	return &PostgresTaskCommentRepository{pool: pool}
}

func (r *PostgresTaskCommentRepository) Create(ctx context.Context, actorID, projectID, taskID int, body string) (*domain.TaskComment, error) {
	comment := &domain.TaskComment{}
	err := r.pool.QueryRow(ctx, `WITH inserted AS (
		INSERT INTO task_comments(task_id, author_id, body)
		SELECT t.id, $3, $4 FROM tasks AS t
		JOIN project_members AS pm ON pm.project_id = t.project_id AND pm.user_id = $3
		WHERE t.project_id = $1 AND t.id = $2
		RETURNING id, task_id, author_id, body, created_at, updated_at
	), history AS (
		INSERT INTO task_history(task_id, actor_id, action, new_values)
		SELECT task_id, author_id, 'comment_created', jsonb_build_object('comment_id', id, 'body', body)
		FROM inserted RETURNING task_id
	)
	SELECT i.id, i.task_id, i.author_id, u.username, i.body, i.created_at, i.updated_at
	FROM inserted AS i JOIN history AS h ON h.task_id = i.task_id
	JOIN users AS u ON u.id = i.author_id`, projectID, taskID, actorID, body).
		Scan(&comment.ID, &comment.TaskID, &comment.AuthorID, &comment.AuthorName, &comment.Body, &comment.CreatedAt, &comment.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrTaskNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}
	return comment, nil
}

func (r *PostgresTaskCommentRepository) List(ctx context.Context, actorID, projectID, taskID, limit, offset int) ([]*domain.TaskComment, error) {
	var allowed bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM tasks AS t
		JOIN project_members AS pm ON pm.project_id = t.project_id AND pm.user_id = $3
		WHERE t.project_id = $1 AND t.id = $2)`, projectID, taskID, actorID).Scan(&allowed)
	if err != nil {
		return nil, fmt.Errorf("check comment access: %w", err)
	}
	if !allowed {
		return nil, repository.ErrTaskNotFound
	}
	rows, err := r.pool.Query(ctx, `SELECT c.id, c.task_id, c.author_id, u.username, c.body, c.created_at, c.updated_at
		FROM task_comments AS c JOIN users AS u ON u.id = c.author_id
		WHERE c.task_id = $1 ORDER BY c.created_at, c.id LIMIT $2 OFFSET $3`, taskID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()
	comments := make([]*domain.TaskComment, 0)
	for rows.Next() {
		comment := &domain.TaskComment{}
		if err := rows.Scan(&comment.ID, &comment.TaskID, &comment.AuthorID, &comment.AuthorName, &comment.Body, &comment.CreatedAt, &comment.UpdatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, rows.Err()
}

func (r *PostgresTaskCommentRepository) Update(ctx context.Context, actorID, projectID, taskID int, commentID int64, body string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var oldBody string
	err = tx.QueryRow(ctx, `SELECT c.body FROM task_comments AS c
		JOIN tasks AS t ON t.id = c.task_id
		JOIN project_members AS pm ON pm.project_id = t.project_id AND pm.user_id = $4
		WHERE c.id = $1 AND c.task_id = $2 AND t.project_id = $3 AND c.author_id = $4 FOR UPDATE OF c`,
		commentID, taskID, projectID, actorID).Scan(&oldBody)
	if errors.Is(err, pgx.ErrNoRows) {
		return repository.ErrCommentNotFound
	}
	if err != nil {
		return err
	}
	if oldBody == body {
		return nil
	}
	if _, err := tx.Exec(ctx, `UPDATE task_comments SET body = $1, updated_at = NOW() WHERE id = $2`, body, commentID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO task_history(task_id, actor_id, action, old_values, new_values)
		VALUES ($1, $2, 'edited', jsonb_build_object('comment_id', $3::bigint, 'body', $4::text), jsonb_build_object('comment_id', $3::bigint, 'body', $5::text))`,
		taskID, actorID, commentID, oldBody, body); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PostgresTaskCommentRepository) Delete(ctx context.Context, actorID, projectID, taskID int, commentID int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var oldBody string
	err = tx.QueryRow(ctx, `DELETE FROM task_comments AS c USING tasks AS t, project_members AS pm
		WHERE c.id = $1 AND c.task_id = $2 AND c.author_id = $4
		AND t.id = c.task_id AND t.project_id = $3
		AND pm.project_id = t.project_id AND pm.user_id = $4
		RETURNING c.body`, commentID, taskID, projectID, actorID).Scan(&oldBody)
	if errors.Is(err, pgx.ErrNoRows) {
		return repository.ErrCommentNotFound
	}
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO task_history(task_id, actor_id, action, old_values)
		VALUES ($1, $2, 'comment_deleted', jsonb_build_object('comment_id', $3::bigint, 'body', $4::text))`,
		taskID, actorID, commentID, oldBody); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
