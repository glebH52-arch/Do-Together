package postgres

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"fmt"
	"strings"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.EmailOutboxRepository = (*PostgresEmailOutboxRepository)(nil)

type PostgresEmailOutboxRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresEmailOutboxRepository(pool *pgxpool.Pool) *PostgresEmailOutboxRepository {
	return &PostgresEmailOutboxRepository{
		pool: pool,
	}
}

func (r *PostgresEmailOutboxRepository) Create(ctx context.Context, message *domain.EmailOutboxMessage) error {
	if message == nil {
		return repository.ErrEmailOutboxMessageNil
	}

	if err := ctx.Err(); err != nil {
		return err
	}
	executor := executorFromContext(ctx, r.pool)
	err := executor.QueryRow(
		ctx, stmtCreateEmailOutbox,
		message.UserID,
		message.Recipient,
		message.Kind,
		message.Payload,
		message.NextAttemptAt,
		message.CreatedAt).Scan(&message.ID)
	if err != nil {
		return fmt.Errorf("create email outbox : %w", err)
	}
	return nil
}
func (r *PostgresEmailOutboxRepository) ClaimBatch(ctx context.Context, workerID string, now time.Time, leaseExpiredBefore time.Time, limit int) ([]*domain.EmailOutboxMessage, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	workerID = strings.TrimSpace(workerID)
	if workerID == "" {
		return nil, repository.ErrEmailWorkerIDEmpty
	}

	if limit <= 0 || limit > 100 {
		return nil, repository.ErrEmailBatchLimitInvalid
	}
	rows, err := r.pool.Query(
		ctx,
		stmtEmailClaimBatch,
		now,
		leaseExpiredBefore,
		limit,
		workerID,
	)
	if err != nil {
		return nil, fmt.Errorf("claim email outbox batch: %w", err)
	}
	defer rows.Close()

	messages := make([]*domain.EmailOutboxMessage, 0, limit)
	for rows.Next() {
		message := domain.EmailOutboxMessage{}
		if err := rows.Scan(
			&message.ID,
			&message.UserID,
			&message.Recipient,
			&message.Kind,
			&message.Payload,
			&message.DeliveryAttempts,
			&message.NextAttemptAt,
			&message.SentAt,
			&message.CreatedAt,
			&message.LockedAt,
			&message.LockedBy,
			&message.LastError,
		); err != nil {
			return nil, fmt.Errorf("scan claimed email outbox message: %w", err)
		}

		messages = append(messages, &message)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate claimed email outbox messages: %w", err)
	}

	return messages, nil
}
func (r *PostgresEmailOutboxRepository) MarkSent(ctx context.Context, messageID uuid.UUID, workerID string, sentAt time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	workerID = strings.TrimSpace(workerID)

	if messageID == uuid.Nil() {
		return repository.ErrEmailOutboxMessageNotFound
	}
	if workerID == "" {
		return repository.ErrEmailWorkerIDEmpty
	}

	tag, err := r.pool.Exec(
		ctx,
		stmtEmailMarkSent,
		messageID,
		workerID,
		sentAt,
	)
	if err != nil {
		return fmt.Errorf("mark email outbox message sent: %w", err)
	}

	if tag.RowsAffected() != 1 {
		return repository.ErrEmailOutboxMessageNotFound
	}

	return nil
}
func (r *PostgresEmailOutboxRepository) MarkFailed(ctx context.Context, messageID uuid.UUID, workerID string, nextAttemptAt time.Time, lastError string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	workerID = strings.TrimSpace(workerID)
	lastError = strings.TrimSpace(lastError)

	if messageID == uuid.Nil() {
		return repository.ErrEmailOutboxMessageNotFound
	}
	if workerID == "" {
		return repository.ErrEmailWorkerIDEmpty
	}
	if lastError == "" {
		lastError = "unknown email delivery error"
	}

	tag, err := r.pool.Exec(
		ctx,
		stmtEmailMarkFailed,
		messageID,
		workerID,
		nextAttemptAt,
		lastError,
	)
	if err != nil {
		return fmt.Errorf("mark email outbox message failed: %w", err)
	}

	if tag.RowsAffected() != 1 {
		return repository.ErrEmailOutboxMessageNotFound
	}

	return nil

}
