package service

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
)

type NotificationService struct {
	repository repository.NotificationRepository
}

func NewNotificationService(repository repository.NotificationRepository) *NotificationService {
	return &NotificationService{repository: repository}
}

func (s *NotificationService) List(ctx context.Context, userID, limit, offset int) ([]*domain.Notification, error) {
	if userID <= 0 {
		return nil, repository.ErrUserNotFound
	}
	if limit <= 0 || limit > 100 || offset < 0 {
		return nil, repository.ErrInvalidTaskSearch
	}
	return s.repository.List(ctx, userID, limit, offset)
}

func (s *NotificationService) MarkRead(ctx context.Context, userID int, notificationID int64) error {
	if userID <= 0 {
		return repository.ErrUserNotFound
	}
	if notificationID <= 0 {
		return repository.ErrNotificationNotFound
	}
	return s.repository.MarkRead(ctx, userID, notificationID)
}
