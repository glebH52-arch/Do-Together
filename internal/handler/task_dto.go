package handler

import (
	"do-together/internal/domain"
	"encoding/json"
	"errors"
	"time"
	"uuid"
)

type taskRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	DueAt       *time.Time `json:"due_at,omitempty"`
	Type        *string    `json:"type,omitempty"`
	Priority    *string    `json:"priority,omitempty"`
	AssigneeID  *uuid.UUID `json:"assignee_id,omitempty"`
}

type taskUpdateRequest struct {
	Title       *string         `json:"title,omitempty"`
	Description *string         `json:"description,omitempty"`
	Status      *string         `json:"status,omitempty"`
	Type        *string         `json:"type,omitempty"`
	Priority    *string         `json:"priority,omitempty"`
	AssigneeID  json.RawMessage `json:"assignee_id"`
	DueAt       json.RawMessage `json:"due_at"`
}

func parseOptionalUUID(raw json.RawMessage) (domain.OptionalUUID, error) {
	if raw == nil {
		return domain.OptionalUUID{}, nil
	}
	if string(raw) == "null" {
		return domain.OptionalUUID{Set: true}, nil
	}
	var value uuid.UUID
	if err := json.Unmarshal(raw, &value); err != nil {
		return domain.OptionalUUID{}, err
	}
	return domain.OptionalUUID{Set: true, Value: &value}, nil
}

func parseOptionalTime(raw json.RawMessage) (domain.OptionalTime, error) {
	if raw == nil {
		return domain.OptionalTime{}, nil
	}
	if string(raw) == "null" {
		return domain.OptionalTime{Set: true}, nil
	}
	var value time.Time
	if err := json.Unmarshal(raw, &value); err != nil {
		return domain.OptionalTime{}, err
	}
	if value.IsZero() {
		return domain.OptionalTime{}, errors.New("invalid due_at")
	}
	return domain.OptionalTime{Set: true, Value: &value}, nil
}

type boardResponse struct {
	Todo       []taskResponse `json:"todo"`
	InProgress []taskResponse `json:"in_progress"`
	Review     []taskResponse `json:"review"`
	Done       []taskResponse `json:"done"`
}

type taskResponse struct {
	ID                 uuid.UUID           `json:"id"`
	Title              string              `json:"title"`
	Description        string              `json:"description"`
	Type               string              `json:"type"`
	Priority           string              `json:"priority"`
	AssigneeID         *uuid.UUID          `json:"assignee_id"`
	AssigneeName       *string             `json:"assignee_name"`
	AssigneeAvatarURL  *string             `json:"assignee_avatar_url"`
	CreatedBy          uuid.UUID           `json:"created_by"`
	CreatorName        string              `json:"creator_name"`
	CreatorAvatarURL   *string             `json:"creator_avatar_url"`
	Status             string              `json:"status"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          *time.Time          `json:"update_at,omitempty"`
	DueAt              *time.Time          `json:"due_at,omitempty"`
	CompletedAt        *time.Time          `json:"completed_at,omitempty"`
	AllowedTransitions []domain.TaskStatus `json:"allowed_transitions"`
}

func taskToResponse(task *domain.Task) taskResponse {
	return taskResponse{
		ID:                 task.ID,
		Title:              task.Title,
		Description:        task.Description,
		Type:               task.Type,
		Priority:           task.Priority,
		AssigneeID:         task.AssigneeID,
		AssigneeName:       task.AssigneeName,
		AssigneeAvatarURL:  relativeAvatarURL(task.AssigneeAvatarFilename, task.AssigneeAvatarUpdatedAt),
		CreatedBy:          task.CreatedBy,
		CreatorName:        task.CreatorName,
		CreatorAvatarURL:   relativeAvatarURL(task.CreatorAvatarFilename, task.CreatorAvatarUpdatedAt),
		Status:             string(task.Status),
		CreatedAt:          task.CreatedAt,
		UpdatedAt:          task.UpdatedAt,
		DueAt:              task.DueAt,
		CompletedAt:        task.CompletedAt,
		AllowedTransitions: domain.AllowedTaskTransitions(task.Status),
	}
}
