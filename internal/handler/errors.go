package handler

import (
	"do-together/internal/domain"
	"do-together/internal/repository"
	"do-together/internal/service"
	"errors"
	"net/http"
)

func statusFromError(err error) int {

	switch {
	case errors.Is(err, domain.ErrTitleEmpty):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrTitleTooLong):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrGoalEmpty):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrGoalTooLong):
		return http.StatusBadRequest
	case errors.Is(err, repository.ErrProjectNotFound):
		return http.StatusNotFound
	case errors.Is(err, repository.ErrUserEmailAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, repository.ErrUsernameAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, repository.ErrUserNotFound):
		return http.StatusNotFound
	case errors.Is(err, service.ErrPasswordEmpty):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrUsernameEmpty):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrEmailEmpty):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrEmailInvalid):
		return http.StatusBadRequest
	case errors.Is(err, service.ErrInvalidCredentials):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
