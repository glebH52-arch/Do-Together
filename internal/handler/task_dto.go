package handler

import (
	"do-together/internal/domain"
	"time"
)

type taskRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	DueAt       *time.Time `json:"due_at,omitempty"`
}

type taskUpdateRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
}

type taskResponse struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	CreatedBy   int        `json:"created_by"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"update_at,omitempty"`
	DueAt       *time.Time `json:"due_at,omitempty"`
}

func taskToResponse(task *domain.Task) taskResponse {
	return taskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		CreatedBy:   task.CreatedBy,
		Status:      string(task.Status),
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		DueAt:       task.DueAt,
	}
}
