package handler

import (
	"do-together/internal/domain"
	"time"
	"uuid"
)

type requestProjectMember struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
}
type updateProjectMemberRoleRequest struct {
	Role string `json:"role"`
}

type responseProjectMember struct {
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url"`
	Email       string    `json:"email"`
	UserID      uuid.UUID `json:"user_id"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}

func projectMemberToResponse(projectMember *domain.ProjectMember) responseProjectMember {
	return responseProjectMember{
		Username:    projectMember.Username,
		DisplayName: projectMember.DisplayName,
		AvatarURL:   relativeAvatarURL(projectMember.AvatarFilename, projectMember.AvatarUpdatedAt),
		Email:       projectMember.Email,
		UserID:      projectMember.UserID,
		Role:        string(projectMember.Role),
		JoinedAt:    projectMember.JoinedAt,
	}
}
