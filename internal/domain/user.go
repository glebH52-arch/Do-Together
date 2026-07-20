package domain

import (
	"errors"
	"net/mail"
	"strings"
	"time"
)

type User struct {
	ID           int
	Username     string
	Email        string
	PasswordHash string
	Status       UserStatus
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

type UserStatus string

const (
	UserStatusActive UserStatus = "active"
)

var (
	ErrEmailInvalid      = errors.New("email is invalid")
	ErrUsernameEmpty     = errors.New("username is empty")
	ErrEmailEmpty        = errors.New("email is empty")
	ErrPasswordHashEmpty = errors.New("password hash is empty")
)

func validateUsername(username string) (string, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return "", ErrUsernameEmpty
	}
	return username, nil
}

func validateEmail(email string) (string, error) {
	email = strings.TrimSpace(email)
	email = strings.ToLower(email)
	if email == "" {
		return "", ErrEmailEmpty
	}
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return "", ErrEmailInvalid
	}
	return email, nil
}

func NewUser(username string, email string, passwordHash string) (*User, error) {
	username, err := validateUsername(username)
	if err != nil {
		return nil, err
	}
	email, err = validateEmail(email)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(passwordHash) == "" {
		return nil, ErrPasswordHashEmpty
	}

	user := User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Status:       UserStatusActive,
		CreatedAt:    time.Now(),
	}

	return &user, nil
}
