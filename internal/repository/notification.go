package repository

import (
	"context"
	"do-together/internal/domain"
	"errors"
	"uuid"
)

var ErrNotificationNotFound = errors.New("notification not found")

type NotificationRepository interface {
	List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Notification, error)
	CountUnread(ctx context.Context, userID uuid.UUID) (int, error)
	MarkRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error
}
