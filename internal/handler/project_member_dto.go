package handler

import (
	"do-together/internal/domain"
	"time"
)

type requestProjectMember struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
}
type updateProjectMemberRoleRequest struct {
	Role string `json:"role"`
}

type responseProjectMember struct {
	Username string    `json:"username"`
	Email    string    `json:"email"`
	UserID   int       `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

func projectMemberToResponse(projectMember *domain.ProjectMember) responseProjectMember {
	return responseProjectMember{
		Username: projectMember.Username,
		Email:    projectMember.Email,
		UserID:   projectMember.UserID,
		Role:     string(projectMember.Role),
		JoinedAt: projectMember.JoinedAt,
	}
}
