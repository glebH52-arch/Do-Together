package postgres

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"errors"
	"os"
	"testing"
	"time"
	"uuid"
)

func TestTaskDetailsAndHistory(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	projectID, projectErr := testUUID("TEST_PROJECT_ID")
	actorID, actorErr := testUUID("TEST_ACTOR_ID")
	if url == "" || projectErr != nil || actorErr != nil {
		t.Skip("disposable PostgreSQL fixture is not configured")
	}
	ctx := context.Background()
	pool, err := NewPool(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	repo := NewTaskRepository(pool)
	due := time.Now().Add(24 * time.Hour)
	task, err := domain.NewTask(domain.NewTaskParams{
		CreatedBy:   actorID,
		ProjectID:   projectID,
		Title:       "Details test",
		Description: "Details description",
		DueAt:       &due,
	})
	if err != nil {
		t.Fatal(err)
	}
	task.Type, task.Priority, task.AssigneeID = "bug", "high", &actorID
	task, err = repo.Create(ctx, task)
	if err != nil {
		t.Fatal(err)
	}
	if task.AssigneeName == nil || *task.AssigneeName == "" {
		t.Fatal("assignee name missing")
	}
	before := *task
	priority, taskType, status := "critical", "feature", domain.TaskStatusInProgress
	changed, err := task.ApplyChanges(domain.TaskChanges{Priority: &priority, Type: &taskType, Status: &status,
		AssigneeID: domain.OptionalUUID{Set: true}, DueAt: domain.OptionalTime{Set: true}})
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("task changes were not detected")
	}
	if err := repo.Update(ctx, &before, task, actorID, projectID); err != nil {
		t.Fatal(err)
	}
	loaded, err := repo.GetByID(ctx, actorID, projectID, task.ID)
	if err != nil || loaded.Type != taskType || loaded.Priority != priority || loaded.AssigneeID != nil || loaded.DueAt != nil || loaded.Status != status {
		t.Fatalf("task details not saved: %+v %v", loaded, err)
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM task_history WHERE task_id=$1 AND action = 'task.updated'`, task.ID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("history: %d %v", count, err)
	}
	invalidAssignee := uuid.New()
	before = *loaded
	loaded.AssigneeID = &invalidAssignee
	if err := repo.Update(ctx, &before, loaded, actorID, projectID); !errors.Is(err, repository.ErrAssigneeNotMember) {
		t.Fatalf("invalid assignee: %v", err)
	}
}
