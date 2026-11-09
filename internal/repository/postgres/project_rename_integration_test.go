package postgres

import (
	"context"
	"do-together/internal/repository"
	"do-together/internal/service"
	"errors"
	"os"
	"strconv"
	"testing"
)

func TestProjectRenamePreservesTaskKeys(t *testing.T) {
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
	projects := NewPostgresProjectRepository(pool)
	before, err := projects.GetByID(ctx, actorID, projectID)
	if err != nil {
		t.Fatal(err)
	}
	var taskID int
	var key string
	if err := pool.QueryRow(ctx, `SELECT id, task_key FROM tasks WHERE project_id=$1 ORDER BY created_at LIMIT 1`, projectID).Scan(&taskID, &key); err != nil {
		t.Fatal(err)
	}
	original := before.Title
	t.Cleanup(func() {
		_, _ = service.NewProjectService(projects).UpdateProject(context.Background(), actorID, projectID, &original, nil)
	})
	changed := original + " renamed"
	if _, err := service.NewProjectService(projects).UpdateProject(ctx, actorID, projectID, &changed, nil); err != nil {
		t.Fatal(err)
	}
	after, err := projects.GetByID(ctx, actorID, projectID)
	if err != nil || after.Title != changed {
		t.Fatalf("rename did not persist: %+v %v", after, err)
	}
	var currentKey string
	if err := pool.QueryRow(ctx, `SELECT task_key FROM tasks WHERE id=$1`, taskID).Scan(&currentKey); err != nil || currentKey != key {
		t.Fatalf("task key changed: %q -> %q, %v", key, currentKey, err)
	}
	if _, err := service.NewProjectService(projects).UpdateProject(ctx, actorID+1000000, projectID, &original, nil); !errors.Is(err, repository.ErrProjectNotFound) {
		t.Fatalf("unrelated user changed project: %v", err)
	}
}
