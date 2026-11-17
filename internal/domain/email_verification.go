package domain

import (
	"errors"
	"strings"
	"time"
	"uuid"
)

type EmailVerificationCode struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	CodeHash       string
	ExpiresAt      time.Time
	ConsumedAt     *time.Time
	FailedAttempts int
	CreatedAt      time.Time
}

var (
	ErrCodeHashEmpty                = errors.New("codehash is empty")
	ErrTTLRequired                  = errors.New("ttl is recuired")
	ErrVerificationCodeConsumed     = errors.New("verification code consumed")
	ErrVerificationCodeExpired      = errors.New("verification code expired")
	ErrVerificationAttemptsExceeded = errors.New("attempts exceeded")
)

const MaxVerificationAttempts = 5

func NewEmailVerificationCode(userID uuid.UUID, codeHash string, now time.Time, ttl time.Duration) (*EmailVerificationCode, error) {
	if userID == uuid.Nil() {
		return nil, ErrInvalidUserID
	}
	codeHash = strings.TrimSpace(codeHash)
	if codeHash == "" {
		return nil, ErrCodeHashEmpty
	}
	if ttl <= 0 {
		return nil, ErrTTLRequired
	}

	return &EmailVerificationCode{
		UserID:         userID,
		CodeHash:       codeHash,
		ExpiresAt:      now.Add(ttl),
		CreatedAt:      now,
		FailedAttempts: 0,
		ConsumedAt:     nil,
	}, nil
}

func (c *EmailVerificationCode) ValidateAttempt(now time.Time) error {
	if c.ConsumedAt != nil {
		return ErrVerificationCodeConsumed
	}
	if !now.Before(c.ExpiresAt) {
		return ErrVerificationCodeExpired
	}
	if c.FailedAttempts >= MaxVerificationAttempts {
		return ErrVerificationAttemptsExceeded
	}
	return nil
}

func (c *EmailVerificationCode) RegisterFailedAttempt() error {
	if c.FailedAttempts >= MaxVerificationAttempts {
		return ErrVerificationAttemptsExceeded
	}
	c.FailedAttempts++

	return nil
}

func (c *EmailVerificationCode) Consume(now time.Time) error {
	err := c.ValidateAttempt(now)

	if err != nil {
		return err
	}

	var value *time.Time
	value = &now
	c.ConsumedAt = value

	return nil
}
