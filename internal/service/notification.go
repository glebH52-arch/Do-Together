package service

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"uuid"
)

type NotificationService struct {
	repository repository.NotificationRepository
}

func NewNotificationService(repository repository.NotificationRepository) *NotificationService {
	return &NotificationService{repository: repository}
}

func (s *NotificationService) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Notification, error) {
	if userID == uuid.Nil() {
		return nil, repository.ErrUserNotFound
	}
	if limit <= 0 || limit > 100 || offset < 0 {
		return nil, repository.ErrInvalidTaskSearch
	}
	return s.repository.List(ctx, userID, limit, offset)
}

func (s *NotificationService) MarkRead(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error {
	if userID == uuid.Nil() {
		return repository.ErrUserNotFound
	}
	if notificationID == uuid.Nil() {
		return repository.ErrNotificationNotFound
	}
	return s.repository.MarkRead(ctx, userID, notificationID)
}

func (s *NotificationService) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	if userID == uuid.Nil() {
		return 0, repository.ErrUserNotFound
	}
	return s.repository.CountUnread(ctx, userID)
}
