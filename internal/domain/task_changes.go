package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidTaskType     = errors.New("invalid task type")
	ErrInvalidTaskPriority = errors.New("invalid task priority")
	ErrInvalidAssigneeID   = errors.New("invalid assignee ID")
)

type OptionalInt struct {
	Set   bool
	Value *int
}
type OptionalTime struct {
	Set   bool
	Value *time.Time
}

type TaskChanges struct {
	Title       *string
	Description *string
	Status      *TaskStatus
	Type        *string
	Priority    *string
	AssigneeID  OptionalInt
	DueAt       OptionalTime
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

func sameInt(left, right *int) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func sameTime(left, right *time.Time) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return left.Equal(*right)
}

func (t *Task) ApplyChanges(changes TaskChanges) error {
	if changes.Type != nil {
		if err := ValidateTaskType(*changes.Type); err != nil {
			return err
		}
	}
	if changes.Priority != nil {
		if err := ValidateTaskPriority(*changes.Priority); err != nil {
			return err
		}
	}
	if changes.AssigneeID.Set && changes.AssigneeID.Value != nil && *changes.AssigneeID.Value <= 0 {
		return ErrInvalidAssigneeID
	}
	oldUpdatedAt := t.UpdatedAt
	if err := t.Update(changes.Title, changes.Description, changes.Status); err != nil {
		return err
	}
	extraChanged := false
	if changes.Type != nil && t.Type != *changes.Type {
		t.Type = *changes.Type
		extraChanged = true
	}
	if changes.Priority != nil && t.Priority != *changes.Priority {
		t.Priority = *changes.Priority
		extraChanged = true
	}
	if changes.AssigneeID.Set && !sameInt(t.AssigneeID, changes.AssigneeID.Value) {
		t.AssigneeID = changes.AssigneeID.Value
		extraChanged = true
	}
	if changes.DueAt.Set && !sameTime(t.DueAt, changes.DueAt.Value) {
		t.DueAt = changes.DueAt.Value
		extraChanged = true
	}
	if extraChanged && t.UpdatedAt == oldUpdatedAt {
		now := time.Now()
		t.UpdatedAt = &now
	}
	return nil
}
