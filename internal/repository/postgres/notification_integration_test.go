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

func TestTaskNotifications(t *testing.T) {
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
	var recipientID uuid.UUID
	if err := pool.QueryRow(ctx, `SELECT id FROM users WHERE id <> $1 AND NOT EXISTS
		(SELECT 1 FROM project_members WHERE project_id = $2 AND user_id = users.id) LIMIT 1`, actorID, projectID).Scan(&recipientID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO project_members(project_id,user_id,role) VALUES ($1,$2,'member')`, projectID, recipientID); err != nil {
		t.Fatal(err)
	}
	defer pool.Exec(ctx, `DELETE FROM project_members WHERE project_id=$1 AND user_id=$2`, projectID, recipientID)
	taskRepo := NewTaskRepository(pool)
	notificationRepo := NewNotificationRepository(pool)
	task, err := domain.NewTask(domain.NewTaskParams{
		CreatedBy:   actorID,
		ProjectID:   projectID,
		Title:       "Notification test",
		Description: "Notification description",
	})
	if err != nil {
		t.Fatal(err)
	}
	task.AssigneeID = &recipientID
	task, err = taskRepo.Create(ctx, task)
	if err != nil {
		t.Fatal(err)
	}
	before := *task
	status := domain.TaskStatusInProgress
	due := time.Now().Add(48 * time.Hour)
	changed, err := task.ApplyChanges(domain.TaskChanges{Status: &status, DueAt: domain.OptionalTime{Set: true, Value: &due}})
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("task changes were not detected")
	}
	if err := taskRepo.Update(ctx, &before, task, actorID, projectID); err != nil {
		t.Fatal(err)
	}
	items, err := notificationRepo.List(ctx, recipientID, 100, 0)
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	var notificationID uuid.UUID
	for _, item := range items {
		if item.TaskID != nil && *item.TaskID == task.ID {
			counts[item.Kind]++
			notificationID = item.ID
		}
	}
	for _, kind := range []string{"assignment_changed", "status_changed", "due_at_changed"} {
		if counts[kind] != 1 {
			t.Fatalf("notification %s count: %d", kind, counts[kind])
		}
	}
	if err := notificationRepo.MarkRead(ctx, actorID, notificationID); !errors.Is(err, repository.ErrNotificationNotFound) {
		t.Fatalf("other user's notification: %v", err)
	}
	if err := notificationRepo.MarkRead(ctx, recipientID, notificationID); err != nil {
		t.Fatal(err)
	}
	items, err = notificationRepo.List(ctx, recipientID, 100, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range items {
		if item.ID == notificationID && item.ReadAt == nil {
			t.Fatal("notification remains unread")
		}
	}
}
