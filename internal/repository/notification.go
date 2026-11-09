package repository

import (
	"context"
	"do-together/internal/domain"
	"errors"
)

var ErrNotificationNotFound = errors.New("notification not found")

type NotificationRepository interface {
	List(ctx context.Context, userID, limit, offset int) ([]*domain.Notification, error)
	MarkRead(ctx context.Context, userID int, notificationID int64) error
}
