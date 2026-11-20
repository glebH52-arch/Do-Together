package domain

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
	"uuid"
)

type Task struct {
	ID                      uuid.UUID
	ProjectID               uuid.UUID
	Title                   string
	Description             string
	Type                    string
	Priority                string
	AssigneeID              *uuid.UUID
	AssigneeName            *string
	AssigneeAvatarFilename  *string
	AssigneeAvatarUpdatedAt *time.Time
	CreatedBy               uuid.UUID
	CreatorName             string
	CreatorAvatarFilename   *string
	CreatorAvatarUpdatedAt  *time.Time
	Status                  TaskStatus
	CreatedAt               time.Time
	UpdatedAt               *time.Time
	DueAt                   *time.Time
	CompletedAt             *time.Time
}

type TaskHistory struct {
	ID         uuid.UUID
	TaskID     uuid.UUID
	ActorID    uuid.UUID
	ActorName  string
	Action     string
	OldValues  []byte
	NewValues  []byte
	OccurredAt time.Time
}

type NewTaskParams struct {
	CreatedBy   uuid.UUID
	ProjectID   uuid.UUID
	Title       string
	Description string
	DueAt       *time.Time
	Type        *string
	Priority    *string
	AssigneeID  *uuid.UUID
}

type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusReview     TaskStatus = "review"
	TaskStatusDone       TaskStatus = "done"
)

var allowedTaskTransitions = map[TaskStatus]map[TaskStatus]bool{
	TaskStatusTodo: {
		TaskStatusInProgress: true,
	},
	TaskStatusInProgress: {
		TaskStatusReview: true,
	},
	TaskStatusReview: {
		TaskStatusInProgress: true,
		TaskStatusDone:       true,
	},
	TaskStatusDone: {
		TaskStatusInProgress: true,
	},
}

var taskStatusOrder = []TaskStatus{
	TaskStatusTodo,
	TaskStatusInProgress,
	TaskStatusReview,
	TaskStatusDone,
}

var (
	ErrTaskDescriptionEmpty   = errors.New("Description empty")
	ErrTaskDescriptionTooLong = errors.New("Description too long")
	ErrInvalidTaskStatus      = errors.New("invalid task status")
	ErrInvalidTaskTransition  = errors.New("invalid task status transition")
)

func validateDescription(description string) (string, error) {
	description = strings.TrimSpace(description)
	if utf8.RuneCountInString(description) > 1000 {
		return "", ErrTaskDescriptionTooLong
	}
	if utf8.RuneCountInString(description) == 0 {
		return "", ErrTaskDescriptionEmpty
	}
	return description, nil
}

func ValidateTaskStatus(status TaskStatus) error {
	switch status {
	case TaskStatusTodo,
		TaskStatusInProgress,
		TaskStatusReview,
		TaskStatusDone:
		return nil
	default:
		return ErrInvalidTaskStatus
	}
}

func ValidateTaskTransition(from, to TaskStatus) error {
	if !allowedTaskTransitions[from][to] {
		return ErrInvalidTaskTransition
	}

	return nil
}

func AllowedTaskTransitions(status TaskStatus) []TaskStatus {
	allowed := allowedTaskTransitions[status]
	transitions := make([]TaskStatus, 0, len(allowed))
	for _, candidate := range taskStatusOrder {
		if allowed[candidate] {
			transitions = append(transitions, candidate)
		}
	}
	return transitions
}

func ValidateTaskType(value string) error {
	switch value {
	case "task", "bug", "feature":
		return nil
	default:
		return ErrInvalidTaskType
	}
}

func ValidateTaskPriority(value string) error {
	switch value {
	case "low", "medium", "high", "critical":
		return nil
	default:
		return ErrInvalidTaskPriority
	}
}

func NewTask(params NewTaskParams) (*Task, error) {
	if params.CreatedBy == uuid.Nil() {
		return nil, ErrInvalidUserID
	}
	if params.ProjectID == uuid.Nil() {
		return nil, ErrInvalidProjectID
	}
	taskTitle, err := validateTitle(params.Title)
	if err != nil {
		return nil, err
	}
	taskDescription, err := validateDescription(params.Description)
	if err != nil {
		return nil, err
	}
	var taskDueAt *time.Time
	if params.DueAt != nil {
		value := *params.DueAt
		taskDueAt = &value
	}

	task := &Task{
		CreatedBy:   params.CreatedBy,
		ProjectID:   params.ProjectID,
		Title:       taskTitle,
		Description: taskDescription,
		Type:        "task",
		Priority:    "medium",
		Status:      TaskStatusTodo,
		CreatedAt:   time.Now(),
		DueAt:       taskDueAt,
	}
	if params.Type != nil {
		if err := ValidateTaskType(*params.Type); err != nil {
			return nil, err
		}
		task.Type = *params.Type
	}
	if params.Priority != nil {
		if err := ValidateTaskPriority(*params.Priority); err != nil {
			return nil, err
		}
		task.Priority = *params.Priority
	}
	if params.AssigneeID != nil {
		if *params.AssigneeID == uuid.Nil() {
			return nil, ErrInvalidAssigneeID
		}

		value := *params.AssigneeID
		task.AssigneeID = &value
	}
	return task, nil
}

func (t *Task) Update(title, description *string, status *TaskStatus) error {
	newTitle := t.Title
	newDescription := t.Description
	newStatus := t.Status

	if title != nil {
		value, err := validateTitle(*title)
		if err != nil {
			return err
		}
		newTitle = value
	}

	if description != nil {
		value, err := validateDescription(*description)
		if err != nil {
			return err
		}
		newDescription = value
	}

	if status != nil {
		if err := ValidateTaskStatus(*status); err != nil {
			return err
		}
		if *status != t.Status {
			if err := ValidateTaskTransition(t.Status, *status); err != nil {
				return err
			}
		}
		newStatus = *status
	}
	if newTitle == t.Title && newDescription == t.Description && newStatus == t.Status {
		return nil
	}
	if newStatus == TaskStatusDone && t.Status != TaskStatusDone {
		now := time.Now()
		t.CompletedAt = &now
	} else if t.Status == TaskStatusDone && newStatus != TaskStatusDone {
		t.CompletedAt = nil
	}
	t.Title = newTitle
	t.Description = newDescription
	t.Status = newStatus

	now := time.Now()
	t.UpdatedAt = &now

	return nil
}
