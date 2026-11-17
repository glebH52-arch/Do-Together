package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"
	"uuid"
)

type EmailKind string

const (
	EmailKindVerification  EmailKind = "email_verification"
	EmailKindPasswordReset EmailKind = "email_password_reset"
	EmailKindProjectInvite EmailKind = "email_project_invite"
)

var (
	ErrInvalidEmailKind    = errors.New("invalid email kind")
	ErrEmailPayloadEmpty   = errors.New("email payload is empty")
	ErrEmailPayloadInvalid = errors.New("email payload is invalid")
)

func ValidateEmailKind(kind EmailKind) error {
	switch kind {
	case EmailKindVerification,
		EmailKindPasswordReset,
		EmailKindProjectInvite:
		return nil
	default:
		return ErrInvalidEmailKind
	}
}

type EmailOutboxMessage struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	Recipient        string
	Kind             EmailKind
	Payload          json.RawMessage
	DeliveryAttempts int
	NextAttemptAt    time.Time
	SentAt           *time.Time
	CreatedAt        time.Time
	LockedAt         *time.Time
	LockedBy         *string
	LastError        *string
}

func NewEmailOutboxMessage(userID uuid.UUID, recipient string, kind EmailKind, payload json.RawMessage, now time.Time) (*EmailOutboxMessage, error) {
	if userID == uuid.Nil() {
		return nil, ErrInvalidUserID
	}
	recipient, err := NormalizeEmail(recipient)
	if err != nil {
		return nil, err
	}
	if err := ValidateEmailKind(kind); err != nil {
		return nil, err
	}
	payload = bytes.TrimSpace(payload)

	if len(payload) == 0 {
		return nil, ErrEmailPayloadEmpty
	}

	if !json.Valid(payload) {
		return nil, ErrEmailPayloadInvalid
	}

	payloadCopy := append(json.RawMessage(nil), payload...)

	return &EmailOutboxMessage{
		UserID:        userID,
		Recipient:     recipient,
		Kind:          kind,
		Payload:       payloadCopy,
		NextAttemptAt: now,
		CreatedAt:     now,
	}, nil

}
