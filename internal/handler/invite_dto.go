package handler

import (
	"do-together/internal/domain"
	"time"
)

type createInviteRequest struct {
	InviteeID int    `json:"invitee_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
}

type inviteResponse struct {
	ID           int        `json:"id"`
	ProjectID    int        `json:"project_id"`
	InviterID    int        `json:"inviter_id"`
	InviteeID    int        `json:"invitee_id"`
	ProjectTitle string     `json:"project_title"`
	InviterName  string     `json:"inviter_name"`
	Role         string     `json:"role"`
	Status       string     `json:"status"`
	ExpiresAt    time.Time  `json:"expires_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

func inviteToResponse(invite *domain.Invite) inviteResponse {
	return inviteResponse{
		ID:           invite.ID,
		ProjectID:    invite.ProjectID,
		InviterID:    invite.InviterID,
		InviteeID:    invite.InviteeID,
		ProjectTitle: invite.ProjectTitle,
		InviterName:  invite.InviterName,
		Role:         string(invite.Role),
		Status:       string(invite.Status),
		ExpiresAt:    invite.ExpiresAt,
		CreatedAt:    invite.CreatedAt,
		UpdatedAt:    invite.UpdatedAt,
	}
}
