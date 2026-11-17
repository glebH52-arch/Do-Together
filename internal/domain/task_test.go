package domain

import (
	"errors"
	"testing"
	"uuid"
)

func TestTaskStatusTransitionsAndCompletion(t *testing.T) {
	task, err := NewTask(NewTaskParams{
		CreatedBy:   uuid.New(),
		ProjectID:   uuid.New(),
		Title:       "Task",
		Description: "Description",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, next := range []TaskStatus{TaskStatusInProgress, TaskStatusReview, TaskStatusDone} {
		if err := task.Update(nil, nil, &next); err != nil {
			t.Fatalf("transition to %s: %v", next, err)
		}
	}
	if task.CompletedAt == nil {
		t.Fatal("done task has no completion time")
	}
	next := TaskStatusInProgress
	if err := task.Update(nil, nil, &next); err != nil {
		t.Fatal(err)
	}
	if task.CompletedAt != nil {
		t.Fatal("reopened task retains completion time")
	}
	previousUpdate := task.UpdatedAt
	if err := task.Update(nil, nil, &next); err != nil || task.UpdatedAt != previousUpdate {
		t.Fatalf("no-op changed task: err=%v", err)
	}
}

func TestTaskRejectsInvalidTransition(t *testing.T) {
	task, err := NewTask(NewTaskParams{
		CreatedBy:   uuid.New(),
		ProjectID:   uuid.New(),
		Title:       "Task",
		Description: "Description",
	})
	if err != nil {
		t.Fatal(err)
	}
	next := TaskStatusDone
	if err := task.Update(nil, nil, &next); !errors.Is(err, ErrInvalidTaskTransition) {
		t.Fatalf("todo -> done: %v", err)
	}
	if task.Status != TaskStatusTodo || task.UpdatedAt != nil || task.CompletedAt != nil {
		t.Fatal("rejected transition changed task")
	}
}
