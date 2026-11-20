package repository

import (
	"context"
	"do-together/internal/domain"
	"errors"
	"uuid"
)

var (
	ErrEmailVerificationCodeNil         = errors.New("email verification code is nil")
	ErrEmailVerificationCodeNotFound    = errors.New("email verification code not found")
	ErrEmailVerificationCodeNotConsumed = errors.New("email verification code is not consumed")
)

type EmailVerificationCodeRepository interface {
	Create(ctx context.Context, code *domain.EmailVerificationCode) error
	GetLatestForUpdate(ctx context.Context, userID uuid.UUID) (*domain.EmailVerificationCode, error)
	UpdateFailedAttempts(ctx context.Context, code *domain.EmailVerificationCode) error
	Consume(ctx context.Context, code *domain.EmailVerificationCode) error
}
