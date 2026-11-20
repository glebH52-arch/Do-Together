package postgres

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"errors"
	"os"
	"sync"
	"testing"
)

func TestConcurrentTaskCreation(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	projectID, projectErr := testUUID("TEST_PROJECT_ID")
	actorID, actorErr := testUUID("TEST_ACTOR_ID")
	if projectErr != nil || actorErr != nil {
		t.Skip("TEST_PROJECT_ID and TEST_ACTOR_ID must identify a member of a disposable test project")
	}
	ctx := context.Background()
	pool, err := NewPool(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	repo := NewTaskRepository(pool)

	var tasks [2]*domain.Task
	var errs [2]error
	var wg sync.WaitGroup
	for i := range tasks {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			task, err := domain.NewTask(domain.NewTaskParams{
				CreatedBy:   actorID,
				ProjectID:   projectID,
				Title:       "Parallel task",
				Description: "Concurrent key allocation",
			})
			if err == nil {
				tasks[index], errs[index] = repo.Create(ctx, task)
			} else {
				errs[index] = err
			}
		}(i)
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if tasks[0].ID == tasks[1].ID {
		t.Fatalf("duplicate task identity: %+v %+v", tasks[0], tasks[1])
	}
	for _, task := range tasks {
		loaded, err := repo.GetByID(ctx, actorID, projectID, task.ID)
		if err != nil || loaded.CreatorName == "" {
			t.Fatalf("read task %s: task=%+v err=%v", task.ID, loaded, err)
		}
	}
	before := *tasks[0]
	next := domain.TaskStatusInProgress
	if err := tasks[0].Update(nil, nil, &next); err != nil {
		t.Fatal(err)
	}
	if err := repo.Update(ctx, &before, tasks[0], actorID, projectID); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM task_history WHERE task_id = $1`, tasks[0].ID).Scan(&count); err != nil || count != 2 {
		t.Fatalf("history count: %d, err=%v", count, err)
	}
	if err := repo.Update(ctx, &before, tasks[0], actorID, projectID); !errors.Is(err, repository.ErrTaskConflict) {
		t.Fatalf("stale update: %v", err)
	}
	if _, err := pool.Exec(ctx, `CREATE FUNCTION test_reject_task_history() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'forced history failure'; END $$`); err != nil {
		t.Fatal(err)
	}
	defer pool.Exec(ctx, `DROP FUNCTION IF EXISTS test_reject_task_history() CASCADE`)
	if _, err := pool.Exec(ctx, `CREATE TRIGGER test_reject_task_history BEFORE INSERT ON task_history FOR EACH ROW EXECUTE FUNCTION test_reject_task_history()`); err != nil {
		t.Fatal(err)
	}
	failedTitle := "Task creation rollback"
	failedTask, err := domain.NewTask(domain.NewTaskParams{
		CreatedBy:   actorID,
		ProjectID:   projectID,
		Title:       failedTitle,
		Description: "History insertion must roll back the task",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Create(ctx, failedTask); err == nil {
		t.Fatal("history failure did not roll back task creation")
	}
	var taskCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM tasks WHERE project_id = $1 AND title = $2`, projectID, failedTitle).Scan(&taskCount); err != nil {
		t.Fatal(err)
	}
	if taskCount != 0 {
		t.Fatalf("task remains after creation rollback: %d", taskCount)
	}
	before = *tasks[1]
	if err := tasks[1].Update(nil, nil, &next); err != nil {
		t.Fatal(err)
	}
	if err := repo.Update(ctx, &before, tasks[1], actorID, projectID); err == nil {
		t.Fatal("history failure did not roll back task update")
	}
	loaded, err := repo.GetByID(ctx, actorID, projectID, tasks[1].ID)
	if err != nil || loaded.Status != domain.TaskStatusTodo {
		t.Fatalf("task changed despite history failure: task=%+v err=%v", loaded, err)
	}
}
