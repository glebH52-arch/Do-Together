package postgres

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"errors"
	"os"
	"strconv"
	"sync"
	"testing"
)

func TestConcurrentTaskKeys(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	projectID, projectErr := strconv.Atoi(os.Getenv("TEST_PROJECT_ID"))
	actorID, actorErr := strconv.Atoi(os.Getenv("TEST_ACTOR_ID"))
	if projectErr != nil || actorErr != nil || projectID <= 0 || actorID <= 0 {
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
			task, err := domain.NewTask(actorID, projectID, "Parallel task", "Concurrent key allocation", nil)
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
	if tasks[0].Key == tasks[1].Key || tasks[0].UUID == tasks[1].UUID {
		t.Fatalf("duplicate task identity: %+v %+v", tasks[0], tasks[1])
	}
	for _, task := range tasks {
		loaded, err := repo.GetByID(ctx, actorID, projectID, task.ID)
		if err != nil || loaded.Key != task.Key || loaded.CreatorName == "" {
			t.Fatalf("read task %d: task=%+v err=%v", task.ID, loaded, err)
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

func TestTaskPositions(t *testing.T) {
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
	tasks := make([]*domain.Task, 3)
	for i := range tasks {
		task, err := domain.NewTask(actorID, projectID, "Position test", "Position test description", nil)
		if err == nil {
			tasks[i], err = repo.Create(ctx, task)
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	status := domain.TaskStatusTodo
	initial, err := repo.List(ctx, actorID, projectID)
	if err != nil {
		t.Fatal(err)
	}
	firstID := 0
	for _, task := range initial {
		if task.Status == status {
			firstID = task.ID
			break
		}
	}
	if firstID == 0 {
		t.Fatal("todo column is empty")
	}
	move := func(before *int) {
		t.Helper()
		task, err := repo.GetByID(ctx, actorID, projectID, tasks[2].ID)
		if err != nil {
			t.Fatal(err)
		}
		if err := repo.Move(ctx, actorID, projectID, task.ID, status, task.Position, before); err != nil {
			t.Fatal(err)
		}
	}
	move(&firstID)
	first, _ := repo.GetByID(ctx, actorID, projectID, tasks[2].ID)
	anchor, _ := repo.GetByID(ctx, actorID, projectID, firstID)
	if first.Position >= anchor.Position {
		t.Fatal("move to beginning failed")
	}
	move(nil)
	last, _ := repo.GetByID(ctx, actorID, projectID, tasks[2].ID)
	anchor, _ = repo.GetByID(ctx, actorID, projectID, tasks[1].ID)
	if last.Position <= anchor.Position {
		t.Fatal("move to end failed")
	}
	move(&tasks[1].ID)
	middle, _ := repo.GetByID(ctx, actorID, projectID, tasks[2].ID)
	left, _ := repo.GetByID(ctx, actorID, projectID, tasks[0].ID)
	right, _ := repo.GetByID(ctx, actorID, projectID, tasks[1].ID)
	if middle.Position <= left.Position || middle.Position >= right.Position {
		t.Fatal("move between neighbors failed")
	}
	var results [2]error
	var wg sync.WaitGroup
	for i, before := range []*int{&tasks[0].ID, nil} {
		wg.Add(1)
		go func(index int, anchor *int) {
			defer wg.Done()
			results[index] = repo.Move(ctx, actorID, projectID, tasks[2].ID, status, middle.Position, anchor)
		}(i, before)
	}
	wg.Wait()
	if !((results[0] == nil && errors.Is(results[1], repository.ErrTaskConflict)) ||
		(results[1] == nil && errors.Is(results[0], repository.ErrTaskConflict))) {
		t.Fatalf("concurrent moves: %v, %v", results[0], results[1])
	}
	ordered, err := repo.List(ctx, actorID, projectID)
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < len(ordered); i++ {
		if ordered[i-1].Status == ordered[i].Status && ordered[i-1].Position > ordered[i].Position {
			t.Fatal("task order was not preserved on re-read")
		}
	}
	var outsiderID int
	if err := pool.QueryRow(ctx, `SELECT id FROM users WHERE NOT EXISTS
		(SELECT 1 FROM project_members WHERE project_id = $1 AND user_id = users.id) LIMIT 1`, projectID).Scan(&outsiderID); err == nil {
		if _, err := repo.List(ctx, outsiderID, projectID); !errors.Is(err, repository.ErrProjectNotFound) {
			t.Fatalf("outsider listed board: %v", err)
		}
		if err := repo.Move(ctx, outsiderID, projectID, tasks[2].ID, status, middle.Position, nil); !errors.Is(err, repository.ErrProjectNotFound) {
			t.Fatalf("outsider moved task: %v", err)
		}
	}
}
