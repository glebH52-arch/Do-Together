package postgres

import (
	"context"
	"do-together/internal/repository"
	"do-together/internal/service"
	"errors"
	"os"
	"testing"
)

func TestProjectRename(t *testing.T) {
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
	projects := NewPostgresProjectRepository(pool)
	before, err := projects.GetByID(ctx, actorID, projectID)
	if err != nil {
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
	if _, err := service.NewProjectService(projects).UpdateProject(ctx, testDifferentUUID(actorID), projectID, &original, nil); !errors.Is(err, repository.ErrProjectNotFound) {
		t.Fatalf("unrelated user changed project: %v", err)
	}
}
