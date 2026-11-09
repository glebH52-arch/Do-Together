package postgres

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"errors"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestTaskDetailsAndHistory(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	projectID, projectErr := strconv.Atoi(os.Getenv("TEST_PROJECT_ID"))
	actorID, actorErr := strconv.Atoi(os.Getenv("TEST_ACTOR_ID"))
	if url == "" || projectErr != nil || actorErr != nil || projectID <= 0 || actorID <= 0 {
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
	task, err := domain.NewTask(actorID, projectID, "Details test", "Details description", &due)
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
	if err := task.ApplyChanges(domain.TaskChanges{Priority: &priority, Type: &taskType, Status: &status,
		AssigneeID: domain.OptionalInt{Set: true}, DueAt: domain.OptionalTime{Set: true}}); err != nil {
		t.Fatal(err)
	}
	if err := repo.Update(ctx, &before, task, actorID, projectID); err != nil {
		t.Fatal(err)
	}
	loaded, err := repo.GetByID(ctx, actorID, projectID, task.ID)
	if err != nil || loaded.Type != taskType || loaded.Priority != priority || loaded.AssigneeID != nil || loaded.DueAt != nil || loaded.Status != status {
		t.Fatalf("task details not saved: %+v %v", loaded, err)
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM task_history WHERE task_id=$1 AND action IN ('status_changed','priority_changed','assignee_changed','due_at_changed')`, task.ID).Scan(&count); err != nil || count != 4 {
		t.Fatalf("history: %d %v", count, err)
	}
	invalidAssignee := -1
	before = *loaded
	loaded.AssigneeID = &invalidAssignee
	if err := repo.Update(ctx, &before, loaded, actorID, projectID); !errors.Is(err, repository.ErrAssigneeNotMember) {
		t.Fatalf("invalid assignee: %v", err)
	}
}
