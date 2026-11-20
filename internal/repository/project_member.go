package repository

import (
	"context"
	"do-together/internal/domain"
	"errors"
	"uuid"
)

type ProjectMemberRepository interface {
	List(ctx context.Context, actorID, projectID uuid.UUID) ([]*domain.ProjectMember, error)
	Add(ctx context.Context, actorID uuid.UUID, member *domain.ProjectMember) error
	UpdateRole(ctx context.Context, actorID, projectID, targetUserID uuid.UUID, role domain.ProjectMemberRole) error
	Remove(ctx context.Context, actorID, projectID, targetUserID uuid.UUID) error
}

var (
	ErrProjectMemberNotFound      = errors.New("project member not found")
	ErrProjectMemberAlreadyExists = errors.New("member already exists")
	ErrForbidden                  = errors.New("forbidden")
)
