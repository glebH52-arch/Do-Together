package service

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordEmpty = errors.New("password is empty")
)

type UserService struct {
	repository repository.UserRepository
}

func NewUserService(repository repository.UserRepository) *UserService {
	return &UserService{
		repository: repository,
	}
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
	err = s.repository.Create(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetUser(ctx context.Context, id int) (*domain.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if id <= 0 {
		return nil, repository.ErrUserNotFound
	}
	user, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}
