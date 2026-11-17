package postgres

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"errors"
	"fmt"
	"uuid"

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

func (r *PostgresTaskCommentRepository) Create(ctx context.Context, actorID, projectID, taskID uuid.UUID, body string) (*domain.TaskComment, error) {
	comment := &domain.TaskComment{}
	err := r.pool.QueryRow(ctx, stmtCreateTaskComment, projectID, taskID, actorID, body).
		Scan(&comment.ID, &comment.TaskID, &comment.AuthorID, &comment.AuthorName, &comment.Body, &comment.CreatedAt, &comment.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrTaskNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}
	return comment, nil
}

func (r *PostgresTaskCommentRepository) List(ctx context.Context, actorID, projectID, taskID uuid.UUID, limit, offset int) ([]*domain.TaskComment, error) {
	var allowed bool
	err := r.pool.QueryRow(ctx, stmtCommentAccess, projectID, taskID, actorID).Scan(&allowed)
	if err != nil {
		return nil, fmt.Errorf("check comment access: %w", err)
	}
	if !allowed {
		return nil, repository.ErrTaskNotFound
	}
	rows, err := r.pool.Query(ctx, stmtListTaskComments, taskID, limit, offset)
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

func (r *PostgresTaskCommentRepository) Update(ctx context.Context, actorID, projectID, taskID uuid.UUID, commentID uuid.UUID, body string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var oldBody string
	err = tx.QueryRow(ctx, stmtLockOwnComment,
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
	if _, err := tx.Exec(ctx, stmtUpdateOwnComment, body, commentID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, stmtCommentEditHistory,
		taskID, actorID, commentID, oldBody, body); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PostgresTaskCommentRepository) Delete(ctx context.Context, actorID, projectID, taskID uuid.UUID, commentID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var oldBody string
	err = tx.QueryRow(ctx, stmtDeleteOwnComment, commentID, taskID, projectID, actorID).Scan(&oldBody)
	if errors.Is(err, pgx.ErrNoRows) {
		return repository.ErrCommentNotFound
	}
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, stmtCommentDeleteHistory,
		taskID, actorID, commentID, oldBody); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
