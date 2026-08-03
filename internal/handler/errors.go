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
	case errors.Is(err, repository.ErrProjectMemberNotFound):
		return http.StatusNotFound
	case errors.Is(err, repository.ErrProjectMemberAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, repository.ErrForbidden):
		return 403
	case errors.Is(err, domain.ErrInvalidProjectMemberRole):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrInvalidInviterID):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrInvalidInviteeID):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrInvalidInviteRole):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrInviteNotPending):
		return http.StatusConflict
	case errors.Is(err, domain.ErrInviteExpired):
		return http.StatusConflict
	case errors.Is(err, domain.ErrInviteNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrPendingInviteAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, domain.ErrTaskDescriptionEmpty):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrTaskDescriptionTooLong):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrInvalidTaskStatus):
		return http.StatusBadRequest
	case errors.Is(err, repository.ErrTaskNotFound):
		return http.StatusNotFound
	case errors.Is(err, repository.ErrNilTask):
		return http.StatusInternalServerError

	default:
		return http.StatusInternalServerError
	}
}
