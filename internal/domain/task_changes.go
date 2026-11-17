package domain

import (
	"errors"
	"time"
	"uuid"
)

var (
	ErrInvalidTaskType     = errors.New("invalid task type")
	ErrInvalidTaskPriority = errors.New("invalid task priority")
	ErrInvalidAssigneeID   = errors.New("invalid assignee ID")
)

type OptionalUUID struct {
	Set   bool
	Value *uuid.UUID
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
	AssigneeID  OptionalUUID
	DueAt       OptionalTime
}

func sameUUID(left, right *uuid.UUID) bool {
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

func (t *Task) ApplyChanges(changes TaskChanges) (bool, error) {

	if changes.Type != nil {
		if err := ValidateTaskType(*changes.Type); err != nil {
			return false, err
		}
	}
	if changes.Priority != nil {
		if err := ValidateTaskPriority(*changes.Priority); err != nil {
			return false, err
		}
	}
	if changes.AssigneeID.Set &&
		changes.AssigneeID.Value != nil &&
		*changes.AssigneeID.Value == uuid.Nil() {
		return false, ErrInvalidAssigneeID
	}

	oldTitle := t.Title
	oldDescription := t.Description
	oldStatus := t.Status
	oldType := t.Type
	oldPriority := t.Priority
	oldAssigneeID := t.AssigneeID
	oldDueAt := t.DueAt
	oldUpdatedAt := t.UpdatedAt

	if err := t.Update(changes.Title, changes.Description, changes.Status); err != nil {
		return false, err
	}

	if changes.Type != nil {
		t.Type = *changes.Type
	}
	if changes.Priority != nil {
		t.Priority = *changes.Priority
	}

	if changes.AssigneeID.Set {
		if changes.AssigneeID.Value == nil {
			t.AssigneeID = nil
		} else {
			value := *changes.AssigneeID.Value
			t.AssigneeID = &value
		}
	}

	if changes.DueAt.Set {
		if changes.DueAt.Value == nil {
			t.DueAt = nil
		} else {
			value := *changes.DueAt.Value
			t.DueAt = &value
		}
	}

	changed := t.Title != oldTitle ||
		t.Description != oldDescription ||
		t.Status != oldStatus ||
		t.Type != oldType ||
		t.Priority != oldPriority ||
		!sameUUID(t.AssigneeID, oldAssigneeID) ||
		!sameTime(t.DueAt, oldDueAt)

	if changed && t.UpdatedAt == oldUpdatedAt {
		now := time.Now()
		t.UpdatedAt = &now
	}

	return changed, nil
}
