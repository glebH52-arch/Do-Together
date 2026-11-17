package domain

import (
	"errors"
	"testing"
	"time"
)

func TestVerifiedUser(t *testing.T) {
	user := User{
		Username:  "gleb",
		Email:     "solomko@gmail.com",
		CreatedAt: time.Now(),
	}

	if user.IsEmailVerified() {
		t.Fatal("expected unverified user")
	}

	firstTime := time.Date(2026, 11, 16, 10, 0, 0, 0, time.UTC)

	err := user.VerifyEmail(firstTime)
	if err != nil {
		t.Fatalf("verify email: %v", err)
	}

	if !user.IsEmailVerified() {
		t.Fatal("expected verified user")
	}

	if user.EmailVerifiedAt == nil {
		t.Fatal("expected verification time")
	}

	if !user.EmailVerifiedAt.Equal(firstTime) {
		t.Fatalf(
			"expected verification time %v, got %v",
			firstTime,
			*user.EmailVerifiedAt,
		)
	}

	secondTime := firstTime.Add(time.Hour)

	err = user.VerifyEmail(secondTime)
	if !errors.Is(err, ErrEmailAlreadyVerified) {
		t.Fatalf("expected ErrEmailAlreadyVerified, got %v", err)
	}

	if !user.EmailVerifiedAt.Equal(firstTime) {
		t.Fatalf(
			"verification time changed: expected %v, got %v",
			firstTime,
			*user.EmailVerifiedAt,
		)
	}
}
