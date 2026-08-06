package repository

import (
	"context"
	"do-together/internal/domain"
	"errors"
	"time"
)

var (
	ErrSessionNotFound    = errors.New("refresh session not found")
	ErrSessionExpired     = errors.New("refresh session expired")
	ErrTokenMismatch      = errors.New("refresh token mismatch")
	ErrSessionAlreadyUsed = errors.New("refresh session already used")
	ErrRedisUnavailable   = errors.New("redis unavailable")
)

type RefreshSessionRepository interface {
	Create(ctx context.Context, session *domain.RefreshSession, idleTTL time.Duration) error
	Rotate(ctx context.Context, sessionID, oldHash, newHash string, now time.Time, idleTTL time.Duration) (userID int, err error)
	Delete(ctx context.Context, sessionID string) error
}
