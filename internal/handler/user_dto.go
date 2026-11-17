package handler

import (
	"do-together/internal/domain"
	"fmt"
	"time"
	"uuid"
)

func relativeAvatarURL(filename *string, updatedAt *time.Time) *string {
	if filename == nil {
		return nil
	}
	value := "/avatars/" + *filename
	if updatedAt != nil {
		value += "?v=" + fmt.Sprint(updatedAt.UnixNano())
	}
	return &value
}

type createUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type createUserResponse struct {
	ID          uuid.UUID  `json:"id"`
	Username    string     `json:"username"`
	DisplayName string     `json:"display_name"`
	AvatarURL   *string    `json:"avatar_url"`
	Email       string     `json:"email"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

func userToResponse(user *domain.User, avatarURL *string) createUserResponse {
	return createUserResponse{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		AvatarURL:   avatarURL,
		Email:       user.Email,
		Status:      string(user.Status),
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}
