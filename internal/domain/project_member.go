package domain

import (
	"errors"
	"time"
)

type ProjectMember struct {
	Username  string
	Email     string
	ProjectID int
	UserID    int
	Role      ProjectMemberRole
	JoinedAt  time.Time
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

func NewProjectMember(projectID int, userID int, role ProjectMemberRole) (*ProjectMember, error) {
	if projectID <= 0 {
		return nil, ErrInvalidProjectID
	}
	if userID <= 0 {
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
