package service

import (
	"context"
	"do-together/internal/auth"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	repository repository.UserRepository
	jwtManager *auth.JWTManager
}

func NewAuthService(repository repository.UserRepository, jwtManager *auth.JWTManager) *AuthService {
	return &AuthService{
		repository: repository,
		jwtManager: jwtManager,
	}
}

func (a *AuthService) Login(ctx context.Context, email, password string) (string, int64, error) {
	if err := ctx.Err(); err != nil {
		return "", 0, err
	}

	email = strings.TrimSpace(email)
	email = strings.ToLower(email)
	if email == "" {
		return "", 0, domain.ErrEmailEmpty
	}
	if password == "" {
		return "", 0, ErrPasswordEmpty
	}

	user, err := a.repository.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", 0, ErrInvalidCredentials
		}
		return "", 0, err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	)
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", 0, ErrInvalidCredentials
		}
		return "", 0, fmt.Errorf("compare password hash: %w", err)
	}

	return a.jwtManager.CreateAccessToken(user.ID)
}
