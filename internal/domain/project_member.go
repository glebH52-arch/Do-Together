package domain

import (
	"errors"
	"time"
	"uuid"
)

type ProjectMember struct {
	Username        string
	DisplayName     string
	AvatarFilename  *string
	AvatarUpdatedAt *time.Time
	Email           string
	ProjectID       uuid.UUID
	UserID          uuid.UUID
	Role            ProjectMemberRole
	JoinedAt        time.Time
}

type ProjectMemberRole string

const (
	ProjectMemberRoleCreator ProjectMemberRole = "creator"
	ProjectMemberRoleAdmin   ProjectMemberRole = "admin"
	ProjectMemberRoleMember  ProjectMemberRole = "member"
)

var (
	ErrInvalidProjectID         = errors.New("project id invalid")
	ErrInvalidUserID            = errors.New("user id invalid")
	ErrInvalidProjectMemberRole = errors.New("project member role invalid")
)

func NewProjectMember(projectID uuid.UUID, userID uuid.UUID, role ProjectMemberRole) (*ProjectMember, error) {
	if projectID == uuid.Nil() {
		return nil, ErrInvalidProjectID
	}
	if userID == uuid.Nil() {
		return nil, ErrInvalidUserID
	}
	err := role.IsValid()
	if err != nil {
		return nil, err
	}
	return &ProjectMember{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
		JoinedAt:  time.Now(),
	}, nil
}

func (r ProjectMemberRole) IsValid() error {
	switch r {
	case ProjectMemberRoleCreator,
		ProjectMemberRoleAdmin,
		ProjectMemberRoleMember:

	default:
		return ErrInvalidProjectMemberRole
	}
	return nil
}

func (r ProjectMemberRole) validateInviteRole() error {
	switch r {
	case ProjectMemberRoleMember,
		ProjectMemberRoleAdmin:

	default:
		return ErrInvalidInviteRole
	}
	return nil
}
