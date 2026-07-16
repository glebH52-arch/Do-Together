package handler

import (
	"do-together/internal/domain"
	"do-together/internal/repository"
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
	default:
		return http.StatusInternalServerError
	}
}
