package domain

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

type Task struct {
	ID          int
	ProjectID   int
	Title       string
	Description string
	CreatedBy   int
	Status      TaskStatus
	CreatedAt   time.Time
	UpdatedAt   *time.Time
	DueAt       *time.Time
}

type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
)

var (
	ErrTaskDescriptionEmpty   = errors.New("Description empty")
	ErrTaskDescriptionTooLong = errors.New("Description too long")
	ErrInvalidTaskStatus      = errors.New("invalid task status")
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

func validateStatus(status TaskStatus) error {
	switch status {
	case TaskStatusTodo,
		TaskStatusInProgress,
		TaskStatusDone:
		return nil
	default:
		return ErrInvalidTaskStatus
	}
}

func NewTask(createdBy, projectID int, title, Description string, dueAt *time.Time) (*Task, error) {
	if createdBy <= 0 {
		return nil, ErrInvalidUserID
	}
	if projectID <= 0 {
		return nil, ErrInvalidProjectID
	}
	taskTitle, err := validateTitle(title)
	if err != nil {
		return nil, err
	}
	taskDescription, err := validateDescription(Description)
	if err != nil {
		return nil, err
	}
	var taskDueAt *time.Time
	if dueAt != nil {
		value := *dueAt
		taskDueAt = &value
	}
	return &Task{
		CreatedBy:   createdBy,
		ProjectID:   projectID,
		Title:       taskTitle,
		Description: taskDescription,
		Status:      TaskStatusTodo,
		CreatedAt:   time.Now(),
		DueAt:       taskDueAt,
	}, nil
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
		if err := validateStatus(*status); err != nil {
			return err
		}
		newStatus = *status
	}
	t.Title = newTitle
	t.Description = newDescription
	t.Status = newStatus

	now := time.Now()
	t.UpdatedAt = &now

	return nil
}
