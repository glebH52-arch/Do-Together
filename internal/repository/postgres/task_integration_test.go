package postgres

import (
	"context"
	"do-together/internal/domain"
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
}
