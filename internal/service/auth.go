package service

import (
	"context"
	"do-together/internal/auth"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	userRepository           repository.UserRepository
	jwtManager               *auth.JWTManager
	refreshSessionRepository repository.RefreshSessionRepository
	refreshTokenIdleTTL      time.Duration
	refreshTokenAbsoluteTTL  time.Duration
}

func NewAuthService(repository repository.UserRepository, jwtManager *auth.JWTManager, refreshSessionRepository repository.RefreshSessionRepository, refreshTokenIdleTTL time.Duration, refreshTokenAbsoluteTTL time.Duration) *AuthService {
	return &AuthService{
		userRepository:           repository,
		jwtManager:               jwtManager,
		refreshSessionRepository: refreshSessionRepository,
		refreshTokenIdleTTL:      refreshTokenIdleTTL,
		refreshTokenAbsoluteTTL:  refreshTokenAbsoluteTTL,
	}
}

func (a *AuthService) Login(ctx context.Context, email, password string) (accessToken string, refreshToken string, expiresIn int64, err error) {
	if err := ctx.Err(); err != nil {
		return "", "", 0, err
	}

	email = strings.TrimSpace(email)
	email = strings.ToLower(email)
	if email == "" {
		return "", "", 0, domain.ErrEmailEmpty
	}
	if password == "" {
		return "", "", 0, ErrPasswordEmpty
	}

	user, err := a.userRepository.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", "", 0, ErrInvalidCredentials
		}
		return "", "", 0, err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	)
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", "", 0, ErrInvalidCredentials
		}
		return "", "", 0, fmt.Errorf("compare password hash: %w", err)
	}

	accessToken, _, err = a.jwtManager.CreateAccessToken(user.ID)
	if err != nil {
		return "", "", 0, err
	}
	now := time.Now()

	refreshToken, sessionID, tokenHash, err := auth.GenerateRefreshToken()
	if err != nil {
		return "", "", 0, fmt.Errorf("generate refresh token: %w", err)
	}
	session := &domain.RefreshSession{
		SessionID:         sessionID,
		UserID:            user.ID,
		TokenHash:         tokenHash,
		CreatedAt:         now,
		AbsoluteExpiresAt: now.Add(a.refreshTokenAbsoluteTTL),
	}
	err = a.refreshSessionRepository.Create(ctx, session, a.refreshTokenIdleTTL)
	if err != nil {
		return "", "", 0, err
	}
	return accessToken, refreshToken, expiresIn, nil
}
func (a *AuthService) Refresh(ctx context.Context, oldRefreshToken string) (accessToken string, newRefreshToken string, expiresIn int64, err error) {
	if err := ctx.Err(); err != nil {
		return "", "", 0, err
	}
	sessionID, oldSecret, err := auth.ParseRefreshToken(oldRefreshToken)
	if err != nil {
		return "", "", 0, err
	}
	oldHash := auth.HashRefreshSecret(oldSecret)
	newRefreshToken, newHash, err := auth.GenerateRefreshTokenForSession(sessionID)
	if err != nil {
		return "", "", 0, fmt.Errorf("generate new refresh token: %w", err)
	}
	now := time.Now()
	userID, err := a.refreshSessionRepository.Rotate(ctx, sessionID, oldHash, newHash, now, a.refreshTokenIdleTTL)
	if err != nil {
		return "", "", 0, fmt.Errorf("rotate refresh session: %w", err)
	}
	accessToken, expiresIn, err = a.jwtManager.CreateAccessToken(userID)
	if err != nil {
		return "", "", 0, fmt.Errorf("create access token: %w", err)
	}
	return accessToken, newRefreshToken, expiresIn, nil
}
func (a *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	sessionID, _, err := auth.ParseRefreshToken(refreshToken)
	if err != nil {
		return err
	}

	err = a.refreshSessionRepository.Delete(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("delete refresh session: %w", err)
	}

	return nil
}
