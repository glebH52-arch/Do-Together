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
	"uuid"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrEmailNotVerified        = errors.New("email is not verified")
	ErrInvalidVerificationCode = errors.New("invalid email verification code")
)

type AuthService struct {
	userRepository           repository.UserRepository
	verificationRepository   repository.EmailVerificationCodeRepository
	transactor               repository.Transactor
	emailCodeManager         *auth.EmailCodeManager
	jwtManager               *auth.JWTManager
	refreshSessionRepository repository.RefreshSessionRepository
	refreshTokenIdleTTL      time.Duration
	refreshTokenAbsoluteTTL  time.Duration
}

func NewAuthService(
	userRepository repository.UserRepository,
	verificationRepository repository.EmailVerificationCodeRepository,
	transactor repository.Transactor,
	emailCodeManager *auth.EmailCodeManager,
	jwtManager *auth.JWTManager,
	refreshSessionRepository repository.RefreshSessionRepository,
	refreshTokenIdleTTL time.Duration,
	refreshTokenAbsoluteTTL time.Duration,
) *AuthService {
	return &AuthService{
		userRepository:           userRepository,
		verificationRepository:   verificationRepository,
		transactor:               transactor,
		emailCodeManager:         emailCodeManager,
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

	if !user.IsEmailVerified() {
		return "", "", 0, ErrEmailNotVerified
	}

	return a.issueSession(ctx, user.ID)
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
func (a *AuthService) issueSession(ctx context.Context, userID uuid.UUID) (accessToken string, refreshToken string, expiresIn int64, err error) {
	if err := ctx.Err(); err != nil {
		return "", "", 0, err
	}

	accessToken, expiresIn, err = a.jwtManager.CreateAccessToken(userID)
	if err != nil {
		return "", "", 0, fmt.Errorf("create access token: %w", err)
	}

	now := time.Now()
	refreshToken, sessionID, tokenHash, err := auth.GenerateRefreshToken()
	if err != nil {
		return "", "", 0, fmt.Errorf("generate refresh token: %w", err)
	}

	session := &domain.RefreshSession{
		SessionID:         sessionID,
		UserID:            userID,
		TokenHash:         tokenHash,
		CreatedAt:         now,
		AbsoluteExpiresAt: now.Add(a.refreshTokenAbsoluteTTL),
	}
	if err := a.refreshSessionRepository.Create(ctx, session, a.refreshTokenIdleTTL); err != nil {
		return "", "", 0, err
	}

	return accessToken, refreshToken, expiresIn, nil
}
func validateVerificationCode(code string) (string, error) {
	code = strings.TrimSpace(code)

	if len(code) != 6 {
		return "", ErrInvalidVerificationCode
	}

	for _, symbol := range code {
		if symbol < '0' || symbol > '9' {
			return "", ErrInvalidVerificationCode
		}
	}

	return code, nil
}
func (a *AuthService) VerifyEmail(ctx context.Context, email string, plainCode string) (accessToken string, refreshToken string, expiresIn int64, err error) {
	if err := ctx.Err(); err != nil {
		return "", "", 0, err
	}

	email, err = domain.NormalizeEmail(email)
	if err != nil {
		return "", "", 0, err
	}

	plainCode, err = validateVerificationCode(plainCode)
	if err != nil {
		return "", "", 0, err
	}

	now := time.Now()

	var userID uuid.UUID
	var verificationErr error

	err = a.transactor.WithinTx(
		ctx,
		func(txCtx context.Context) error {
			user, err := a.userRepository.GetByEmail(
				txCtx,
				email,
			)
			if err != nil {
				if errors.Is(err, repository.ErrUserNotFound) {
					return ErrInvalidVerificationCode
				}
				return err
			}

			if user.IsEmailVerified() {
				return domain.ErrEmailAlreadyVerified
			}

			code, err :=
				a.verificationRepository.GetLatestForUpdate(
					txCtx,
					user.ID,
				)
			if err != nil {
				return err
			}

			if err := code.ValidateAttempt(now); err != nil {
				return err
			}

			if !a.emailCodeManager.Verify(
				plainCode,
				code.CodeHash,
			) {
				if err := code.RegisterFailedAttempt(); err != nil {
					return err
				}

				if err := a.verificationRepository.UpdateFailedAttempts(
					txCtx,
					code,
				); err != nil {
					return err
				}

				verificationErr = ErrInvalidVerificationCode
				return nil
			}

			if err := code.Consume(now); err != nil {
				return err
			}

			if err := user.VerifyEmail(now); err != nil {
				return err
			}

			if err := a.verificationRepository.Consume(
				txCtx,
				code,
			); err != nil {
				return err
			}

			if err := a.userRepository.UpdateEmailVerification(
				txCtx,
				user,
			); err != nil {
				return err
			}

			userID = user.ID
			return nil
		},
	)
	if err != nil {
		return "", "", 0, err
	}

	if verificationErr != nil {
		return "", "", 0, verificationErr
	}

	return a.issueSession(ctx, userID)
}
