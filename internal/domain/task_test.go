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

func TestTaskTransitionMatrix(t *testing.T) {
	statuses := []TaskStatus{
		TaskStatusTodo,
		TaskStatusInProgress,
		TaskStatusReview,
		TaskStatusDone,
	}
	allowed := map[TaskStatus]map[TaskStatus]bool{
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

	for _, from := range statuses {
		for _, to := range statuses {
			err := ValidateTaskTransition(from, to)
			if allowed[from][to] && err != nil {
				t.Errorf("%s -> %s rejected: %v", from, to, err)
			}
			if !allowed[from][to] && !errors.Is(err, ErrInvalidTaskTransition) {
				t.Errorf("%s -> %s error = %v, want invalid transition", from, to, err)
			}
		}
	}
}

func TestAllowedTaskTransitions(t *testing.T) {
	tests := []struct {
		status TaskStatus
		want   []TaskStatus
	}{
		{TaskStatusTodo, []TaskStatus{TaskStatusInProgress}},
		{TaskStatusInProgress, []TaskStatus{TaskStatusReview}},
		{TaskStatusReview, []TaskStatus{TaskStatusInProgress, TaskStatusDone}},
		{TaskStatusDone, []TaskStatus{TaskStatusInProgress}},
	}

	for _, test := range tests {
		got := AllowedTaskTransitions(test.status)
		if len(got) != len(test.want) {
			t.Fatalf("%s transitions = %v, want %v", test.status, got, test.want)
		}
		for i := range got {
			if got[i] != test.want[i] {
				t.Errorf("%s transitions = %v, want %v", test.status, got, test.want)
			}
		}
	}
}

func TestRejectedTodoToReviewDoesNotMutateTask(t *testing.T) {
	task, err := NewTask(NewTaskParams{
		CreatedBy:   uuid.New(),
		ProjectID:   uuid.New(),
		Title:       "Task",
		Description: "Description",
	})
	if err != nil {
		t.Fatal(err)
	}
	next := TaskStatusReview
	if err := task.Update(nil, nil, &next); !errors.Is(err, ErrInvalidTaskTransition) {
		t.Fatalf("todo -> review: %v", err)
	}
	if task.Status != TaskStatusTodo || task.UpdatedAt != nil || task.CompletedAt != nil {
		t.Fatal("rejected transition changed task")
	}
}
