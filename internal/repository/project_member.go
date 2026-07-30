package repository

import (
	"context"
	"do-together/internal/domain"
	"errors"
)

type ProjectMemberRepository interface {
	List(ctx context.Context, actorID, projectID int) ([]*domain.ProjectMember, error)
	Add(ctx context.Context, actorID int, member *domain.ProjectMember) error
	UpdateRole(ctx context.Context, actorID, projectID, targetUserID int, role domain.ProjectMemberRole) error
	Remove(ctx context.Context, actorID, projectID, targetUserID int) error
}

var (
	ErrProjectMemberNotFound      = errors.New("project member not found")
	ErrProjectMemberAlreadyExists = errors.New("member already exists")
	ErrForbidden                  = errors.New("forbidden")
)
