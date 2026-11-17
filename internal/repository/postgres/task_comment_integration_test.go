package postgres

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"errors"
	"os"
	"testing"
	"uuid"
)

func TestTaskComments(t *testing.T) {
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
	task, err := domain.NewTask(domain.NewTaskParams{
		CreatedBy:   actorID,
		ProjectID:   projectID,
		Title:       "Comments test",
		Description: "Comments description",
	})
	if err != nil {
		t.Fatal(err)
	}
	task, err = NewTaskRepository(pool).Create(ctx, task)
	if err != nil {
		t.Fatal(err)
	}
	repo := NewTaskCommentRepository(pool)
	comment, err := repo.Create(ctx, actorID, projectID, task.ID, "First comment")
	if err != nil || comment.AuthorName == "" {
		t.Fatalf("create: %+v %v", comment, err)
	}
	comments, err := repo.List(ctx, actorID, projectID, task.ID, 1, 0)
	if err != nil || len(comments) != 1 {
		t.Fatalf("list: %+v %v", comments, err)
	}
	comments, err = repo.List(ctx, actorID, projectID, task.ID, 1, 1)
	if err != nil || len(comments) != 0 {
		t.Fatalf("page: %+v %v", comments, err)
	}
	var outsiderID uuid.UUID
	if err := pool.QueryRow(ctx, `SELECT id FROM users WHERE NOT EXISTS (SELECT 1 FROM project_members WHERE project_id=$1 AND user_id=users.id) LIMIT 1`, projectID).Scan(&outsiderID); err == nil {
		if _, err := repo.List(ctx, outsiderID, projectID, task.ID, 20, 0); !errors.Is(err, repository.ErrTaskNotFound) {
			t.Fatalf("outsider list: %v", err)
		}
		if err := repo.Update(ctx, outsiderID, projectID, task.ID, comment.ID, "Stolen edit"); !errors.Is(err, repository.ErrCommentNotFound) {
			t.Fatalf("outsider update: %v", err)
		}
		if err := repo.Delete(ctx, outsiderID, projectID, task.ID, comment.ID); !errors.Is(err, repository.ErrCommentNotFound) {
			t.Fatalf("outsider delete: %v", err)
		}
	}
	if err := repo.Update(ctx, actorID, projectID, task.ID, comment.ID, "Edited comment"); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(ctx, actorID, projectID, task.ID, comment.ID); err != nil {
		t.Fatal(err)
	}
	comments, err = repo.List(ctx, actorID, projectID, task.ID, 20, 0)
	if err != nil || len(comments) != 0 {
		t.Fatalf("deleted comment remains: %+v %v", comments, err)
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM task_history WHERE task_id=$1 AND action IN ('comment.created','comment.deleted')`, task.ID).Scan(&count); err != nil || count != 2 {
		t.Fatalf("history: %d %v", count, err)
	}
}
