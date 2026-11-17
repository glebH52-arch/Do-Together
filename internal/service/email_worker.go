package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"do-together/internal/email"
	"do-together/internal/repository"
)

var (
	ErrEmailOutboxRepositoryNil = errors.New("email outbox repository is nil")
	ErrEmailSenderNil           = errors.New("email sender is nil")
	ErrEmailWorkerIDEmpty       = errors.New("email worker ID is empty")
	ErrEmailWorkerBatchInvalid  = errors.New("email worker batch size must be between 1 and 100")
	ErrEmailWorkerLeaseInvalid  = errors.New("email worker lease duration must be positive")
)

type EmailWorker struct {
	outbox        repository.EmailOutboxRepository
	sender        email.Sender
	workerID      string
	batchSize     int
	leaseDuration time.Duration
}

func NewEmailWorker(outbox repository.EmailOutboxRepository, sender email.Sender, workerID string, batchSize int, leaseDuration time.Duration) (*EmailWorker, error) {
	if outbox == nil {
		return nil, ErrEmailOutboxRepositoryNil
	}
	if sender == nil {
		return nil, ErrEmailSenderNil
	}

	workerID = strings.TrimSpace(workerID)
	if workerID == "" {
		return nil, ErrEmailWorkerIDEmpty
	}
	if batchSize <= 0 || batchSize > 100 {
		return nil, ErrEmailWorkerBatchInvalid
	}
	if leaseDuration <= 0 {
		return nil, ErrEmailWorkerLeaseInvalid
	}

	return &EmailWorker{
		outbox:        outbox,
		sender:        sender,
		workerID:      workerID,
		batchSize:     batchSize,
		leaseDuration: leaseDuration,
	}, nil
}
func (w *EmailWorker) ProcessBatch(ctx context.Context) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	now := time.Now()
	leaseExpiredBefore := now.Add(-w.leaseDuration)

	messages, err := w.outbox.ClaimBatch(
		ctx,
		w.workerID,
		now,
		leaseExpiredBefore,
		w.batchSize,
	)
	if err != nil {
		return 0, fmt.Errorf("claim email batch: %w", err)
	}

	if len(messages) == 0 {
		return 0, nil
	}

	var processErrors []error
	processed := 0

	for _, message := range messages {
		sendErr := w.sender.Send(ctx, message)
		if sendErr != nil {
			nextAttemptAt := time.Now().Add(
				emailRetryDelay(message.DeliveryAttempts),
			)

			markErr := w.outbox.MarkFailed(
				ctx,
				message.ID,
				w.workerID,
				nextAttemptAt,
				sendErr.Error(),
			)
			if markErr != nil {
				processErrors = append(
					processErrors,
					fmt.Errorf(
						"send email %s: %v; mark failed: %w",
						message.ID,
						sendErr,
						markErr,
					),
				)
				continue
			}

			processed++
			processErrors = append(
				processErrors,
				fmt.Errorf("send email %s: %w", message.ID, sendErr),
			)
			continue
		}

		if err := w.outbox.MarkSent(
			ctx,
			message.ID,
			w.workerID,
			time.Now(),
		); err != nil {
			processErrors = append(
				processErrors,
				fmt.Errorf("mark email %s sent: %w", message.ID, err),
			)
			continue
		}

		processed++
	}

	return processed, errors.Join(processErrors...)

}
func emailRetryDelay(deliveryAttempts int) time.Duration {
	switch {
	case deliveryAttempts <= 0:
		return time.Second * 20
	case deliveryAttempts == 1:
		return 2 * time.Second * 20
	case deliveryAttempts == 2:
		return 4 * time.Second * 20
	case deliveryAttempts == 3:
		return 8 * time.Second * 20
	default:
		return 16 * time.Second * 20
	}
}
