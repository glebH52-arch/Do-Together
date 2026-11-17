package repository

import (
	"context"
	"do-together/internal/domain"
	"errors"
	"uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	UpdateEmailVerification(ctx context.Context, user *domain.User) error
}

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrNilUser                = errors.New("user is nil")
	ErrUserEmailAlreadyExists = errors.New("user email already exists")
	ErrUsernameAlreadyExists  = errors.New("username already exists")
	ErrUserEmailNotVerified   = errors.New("user email is not verified")
)
