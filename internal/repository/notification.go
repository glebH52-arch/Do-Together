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
	MarkRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error
}
