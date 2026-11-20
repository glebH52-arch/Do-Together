package email

import (
	"context"
	"do-together/internal/domain"
)

type Sender interface {
	Send(ctx context.Context, message *domain.EmailOutboxMessage) error
}
