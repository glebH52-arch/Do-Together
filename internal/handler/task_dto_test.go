package handler

import (
	"do-together/internal/domain"
	"testing"
)

func TestTaskResponseIncludesAllowedTransitions(t *testing.T) {
	task := &domain.Task{Status: domain.TaskStatusReview}

	response := taskToResponse(task)

	want := []domain.TaskStatus{domain.TaskStatusInProgress, domain.TaskStatusDone}
	if len(response.AllowedTransitions) != len(want) {
		t.Fatalf("allowed transitions = %v, want %v", response.AllowedTransitions, want)
	}
	for i := range want {
		if response.AllowedTransitions[i] != want[i] {
			t.Fatalf("allowed transitions = %v, want %v", response.AllowedTransitions, want)
		}
	}
}
