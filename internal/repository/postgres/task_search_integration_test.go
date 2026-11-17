package postgres

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
	"uuid"
)

func TestTaskSearch(t *testing.T) {
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
	title := fmt.Sprintf("Needle-%d", time.Now().UnixNano())
	task, err := domain.NewTask(domain.NewTaskParams{
		CreatedBy:   actorID,
		ProjectID:   projectID,
		Title:       title,
		Description: "Searchable description",
		DueAt:       &due,
	})
	if err != nil {
		t.Fatal(err)
	}
	task, err = repo.Create(ctx, task)
	if err != nil {
		t.Fatal(err)
	}
	options := repository.TaskSearchOptions{Query: title, Status: "todo", Priority: "medium", Type: "task", DueBefore: &due, Sort: "created_at", Direction: "desc", Limit: 1}
	tasks, err := repo.Search(ctx, actorID, projectID, options)
	if err != nil || len(tasks) != 1 || tasks[0].ID != task.ID {
		t.Fatalf("search: %+v %v", tasks, err)
	}
	options.Offset = 1
	tasks, err = repo.Search(ctx, actorID, projectID, options)
	if err != nil || len(tasks) != 0 {
		t.Fatalf("page: %+v %v", tasks, err)
	}
	options.Offset = 0
	options.Priority = "high"
	tasks, err = repo.Search(ctx, actorID, projectID, options)
	if err != nil || len(tasks) != 0 {
		t.Fatalf("priority filter: %+v %v", tasks, err)
	}
	var outsiderID uuid.UUID
	if err := pool.QueryRow(ctx, `SELECT id FROM users WHERE NOT EXISTS (SELECT 1 FROM project_members WHERE project_id=$1 AND user_id=users.id) LIMIT 1`, projectID).Scan(&outsiderID); err == nil {
		if _, err := repo.Search(ctx, outsiderID, projectID, options); !errors.Is(err, repository.ErrProjectNotFound) {
			t.Fatalf("outsider search: %v", err)
		}
	}
}
