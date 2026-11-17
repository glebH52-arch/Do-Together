package domain

import (
	"errors"
	"testing"
	"time"
)

func TestEmailVerificationCodeValidateAttempt(t *testing.T) {
	baseTime := time.Date(2026, 11, 16, 10, 0, 0, 0, time.UTC)
	expiresAt := baseTime.Add(15 * time.Minute)
	consumedAt := baseTime.Add(time.Minute)

	tests := []struct {
		name    string
		code    EmailVerificationCode
		now     time.Time
		wantErr error
	}{
		{
			name: "active code",
			code: EmailVerificationCode{
				ExpiresAt: expiresAt,
			},
			now: baseTime,
		},
		{
			name: "expired exactly at expiration time",
			code: EmailVerificationCode{
				ExpiresAt: expiresAt,
			},
			now:     expiresAt,
			wantErr: ErrVerificationCodeExpired,
		},
		{
			name: "consumed code",
			code: EmailVerificationCode{
				ExpiresAt:  expiresAt,
				ConsumedAt: &consumedAt,
			},
			now:     baseTime,
			wantErr: ErrVerificationCodeConsumed,
		},
		{
			name: "attempt limit reached",
			code: EmailVerificationCode{
				ExpiresAt:      expiresAt,
				FailedAttempts: MaxVerificationAttempts,
			},
			now:     baseTime,
			wantErr: ErrVerificationAttemptsExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.code.ValidateAttempt(tt.now)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateAttempt() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
