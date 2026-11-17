package service

import (
	"context"
	"do-together/internal/auth"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"uuid"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordEmpty = errors.New("password is empty")
)

type UserService struct {
	repository               repository.UserRepository
	verificationRepository   repository.EmailVerificationCodeRepository
	emailOutboxRepository    repository.EmailOutboxRepository
	transactor               repository.Transactor
	emailCodeManager         *auth.EmailCodeManager
	emailVerificationCodeTTL time.Duration
}

func NewUserService(
	userRepository repository.UserRepository,
	verificationRepository repository.EmailVerificationCodeRepository,
	emailOutboxRepository repository.EmailOutboxRepository,
	transactor repository.Transactor,
	emailCodeManager *auth.EmailCodeManager,
	emailVerificationCodeTTL time.Duration,
) *UserService {
	return &UserService{
		repository:               userRepository,
		verificationRepository:   verificationRepository,
		emailOutboxRepository:    emailOutboxRepository,
		transactor:               transactor,
		emailCodeManager:         emailCodeManager,
		emailVerificationCodeTTL: emailVerificationCodeTTL,
	}
}

type emailVerificationPayload struct {
	Code     string `json:"code"`
	Username string `json:"username"`
}

func (s *UserService) CreateUser(ctx context.Context, username, email, password string) (*domain.User, error) {
	if password == "" {
		return nil, ErrPasswordEmpty
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	user, err := domain.NewUser(username, email, string(passwordHash))
	if err != nil {
		return nil, err
	}
	plainCode, codeHash, err := s.emailCodeManager.Generate()
	if err != nil {
		return nil, fmt.Errorf(
			"generate email verification code: %w",
			err,
		)
	}

	payload, err := json.Marshal(emailVerificationPayload{
		Code:     plainCode,
		Username: user.Username,
	})
	if err != nil {
		return nil, fmt.Errorf(
			"marshal email verification payload: %w",
			err,
		)
	}

	now := time.Now()

	err = s.transactor.WithinTx(
		ctx,
		func(txCtx context.Context) error {
			if err := s.repository.Create(txCtx, user); err != nil {
				return err
			}

			code, err := domain.NewEmailVerificationCode(
				user.ID,
				codeHash,
				now,
				s.emailVerificationCodeTTL,
			)
			if err != nil {
				return err
			}

			message, err := domain.NewEmailOutboxMessage(
				user.ID,
				user.Email,
				domain.EmailKindVerification,
				json.RawMessage(payload),
				now,
			)
			if err != nil {
				return err
			}

			if err := s.verificationRepository.Create(
				txCtx,
				code,
			); err != nil {
				return err
			}

			if err := s.emailOutboxRepository.Create(
				txCtx,
				message,
			); err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if id == uuid.Nil() {
		return nil, repository.ErrUserNotFound
	}
	user, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}
