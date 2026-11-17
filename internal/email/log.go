package email

import (
	"context"
	"do-together/internal/domain"
	"log"
)

// LogSender is an explicit local-development transport. It never contacts a remote provider.
type LogSender struct{}

func NewLogSender() *LogSender { return &LogSender{} }

func (s *LogSender) Send(ctx context.Context, message *domain.EmailOutboxMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	log.Printf("local email recipient=%s kind=%s payload=%s", message.Recipient, message.Kind, message.Payload)
	return nil
}
