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

type taskMoveRequest struct {
	ExpectedStatus   domain.TaskStatus `json:"expected_status"`
	ExpectedPosition int64             `json:"expected_position"`
	BeforeTaskID     *int              `json:"before_task_id"`
}

type boardResponse struct {
	Todo       []taskResponse `json:"todo"`
	InProgress []taskResponse `json:"in_progress"`
	Review     []taskResponse `json:"review"`
	Done       []taskResponse `json:"done"`
}

type taskResponse struct {
	ID           int        `json:"id"`
	UUID         string     `json:"uuid"`
	Key          string     `json:"key"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Type         string     `json:"type"`
	Priority     string     `json:"priority"`
	AssigneeID   *int       `json:"assignee_id"`
	AssigneeName *string    `json:"assignee_name"`
	CreatedBy    int        `json:"created_by"`
	CreatorName  string     `json:"creator_name"`
	Status       string     `json:"status"`
	Position     int64      `json:"position"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"update_at,omitempty"`
	DueAt        *time.Time `json:"due_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

func taskToResponse(task *domain.Task) taskResponse {
	return taskResponse{
		ID:           task.ID,
		UUID:         task.UUID,
		Key:          task.Key,
		Title:        task.Title,
		Description:  task.Description,
		Type:         task.Type,
		Priority:     task.Priority,
		AssigneeID:   task.AssigneeID,
		AssigneeName: task.AssigneeName,
		CreatedBy:    task.CreatedBy,
		CreatorName:  task.CreatorName,
		Status:       string(task.Status),
		Position:     task.Position,
		CreatedAt:    task.CreatedAt,
		UpdatedAt:    task.UpdatedAt,
		DueAt:        task.DueAt,
		CompletedAt:  task.CompletedAt,
	}
}
