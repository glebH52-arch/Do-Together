package repository

import (
	"context"
	"do-together/internal/domain"
	"errors"
	"time"
	"uuid"
)

var (
	ErrEmailOutboxMessageNil      = errors.New("email outbox message is nil")
	ErrEmailWorkerIDEmpty         = errors.New("email worker ID is empty")
	ErrEmailBatchLimitInvalid     = errors.New("email batch limit must be between 1 and 100")
	ErrEmailOutboxMessageNotFound = errors.New("email outbox message not found or not owned by worker")
)

type EmailOutboxRepository interface {
	Create(ctx context.Context, message *domain.EmailOutboxMessage) error
	ClaimBatch(ctx context.Context, workerID string, now time.Time, leaseExpiredBefore time.Time, limit int) ([]*domain.EmailOutboxMessage, error)
	MarkSent(ctx context.Context, messageID uuid.UUID, workerID string, sentAt time.Time) error
	MarkFailed(ctx context.Context, messageID uuid.UUID, workerID string, nextAttemptAt time.Time, lastError string) error
}
