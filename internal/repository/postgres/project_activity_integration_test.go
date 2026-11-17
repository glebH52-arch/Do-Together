package postgres

import (
	"context"
	"do-together/internal/repository"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
	"uuid"
)

func TestProjectActivityFeed(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := NewPool(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	ownerID := uuid.New()
	removedMemberID := uuid.New()
	projectID := uuid.New()
	otherProjectID := uuid.New()
	for id, username := range map[uuid.UUID]string{ownerID: "activity-owner", removedMemberID: "activity-member"} {
		value := fmt.Sprintf("%s-%s", username, id.String())
		if _, err := pool.Exec(ctx, `INSERT INTO users(id,username,display_name,email,password_hash,status)
			VALUES ($1,$2,$3,$4,'test','active')`, id, value, username, value+"@example.test"); err != nil {
			t.Fatal(err)
		}
		defer pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, id)
	}
	for _, id := range []uuid.UUID{projectID, otherProjectID} {
		if _, err := pool.Exec(ctx, `INSERT INTO projects(id,created_by,title,goal,status)
			VALUES ($1,$2,$3,'activity test','active')`, id, ownerID, "Activity project"); err != nil {
			t.Fatal(err)
		}
		defer pool.Exec(ctx, `DELETE FROM projects WHERE id=$1`, id)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO project_members(project_id,user_id,role,joined_at)
		VALUES ($1,$2,'creator',NOW()),($1,$3,'member',NOW()),($4,$2,'creator',NOW())`, projectID, ownerID, removedMemberID, otherProjectID); err != nil {
		t.Fatal(err)
	}

	taskOneID := uuid.New()
	taskTwoID := uuid.New()
	for id, title := range map[uuid.UUID]string{taskOneID: "First task", taskTwoID: "Second task"} {
		if _, err := pool.Exec(ctx, `INSERT INTO tasks(id,project_id,title,description,created_by,status,task_type,priority,created_at)
			VALUES ($1,$2,$3,'test',$4,'todo','task','medium',NOW())`, id, projectID, title, ownerID); err != nil {
			t.Fatal(err)
		}
	}

	commonTime := time.Now().UTC().Truncate(time.Microsecond)
	oldIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New()}
	insertEvent := func(id, eventProjectID, objectID uuid.UUID, eventType, objectType, title string, occurredAt time.Time, taskID *uuid.UUID) {
		t.Helper()
		_, err := pool.Exec(ctx, `INSERT INTO task_history(
			id,task_id,project_id,actor_id,actor_name,action,object_type,object_id,object_title,new_values,occurred_at
		) VALUES ($1,$2,$3,$4,'Activity owner',$5,$6,$7,$8,'{}',$9)`,
			id, taskID, eventProjectID, ownerID, eventType, objectType, objectID, title, occurredAt)
		if err != nil {
			t.Fatal(err)
		}
	}
	insertEvent(oldIDs[0], projectID, projectID, "project.renamed", "project", "Activity project", commonTime, nil)
	insertEvent(oldIDs[1], projectID, taskOneID, "task.created", "task", "First task", commonTime, &taskOneID)
	insertEvent(oldIDs[2], projectID, taskTwoID, "task.created", "task", "Second task", commonTime, &taskTwoID)
	insertEvent(oldIDs[3], otherProjectID, otherProjectID, "project.created", "project", "Other project", commonTime, nil)

	repo := NewPostgresProjectActivityRepository(pool)
	firstPage, err := repo.List(ctx, ownerID, projectID, nil, 2)
	if err != nil || len(firstPage) != 2 {
		t.Fatalf("first page: %d %v", len(firstPage), err)
	}
	for _, item := range firstPage {
		if item.ProjectID != projectID {
			t.Fatalf("foreign project event leaked: %s", item.ProjectID)
		}
	}
	cursor := &repository.ActivityCursor{OccurredAt: firstPage[1].OccurredAt, ID: firstPage[1].ID}
	newEventID := uuid.New()
	insertEvent(newEventID, projectID, projectID, "project.archived", "project", "Activity project", commonTime.Add(time.Second), nil)
	secondPage, err := repo.List(ctx, ownerID, projectID, cursor, 10)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[uuid.UUID]bool{}
	for _, item := range append(firstPage, secondPage...) {
		if seen[item.ID] {
			t.Fatalf("duplicate event across pages: %s", item.ID)
		}
		seen[item.ID] = true
	}
	if seen[newEventID] {
		t.Fatal("new event inserted between pages must not shift the old page window")
	}
	for _, id := range oldIDs[:3] {
		if !seen[id] {
			t.Fatalf("old event was skipped: %s", id)
		}
	}

	deletedID := uuid.New()
	insertEvent(deletedID, projectID, taskOneID, "task.deleted", "task", "First task", commonTime.Add(2*time.Second), &taskOneID)
	if _, err := pool.Exec(ctx, `DELETE FROM tasks WHERE id=$1`, taskOneID); err != nil {
		t.Fatal(err)
	}
	items, err := repo.List(ctx, ownerID, projectID, nil, 20)
	if err != nil {
		t.Fatal(err)
	}
	var foundDeleted bool
	for _, item := range items {
		if item.ID == deletedID {
			foundDeleted = item.ObjectDeleted && item.ObjectTitle == "First task"
		}
	}
	if !foundDeleted {
		t.Fatal("deleted task activity did not retain its readable snapshot")
	}

	if _, err := pool.Exec(ctx, `DELETE FROM project_members WHERE project_id=$1 AND user_id=$2`, projectID, removedMemberID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.List(ctx, removedMemberID, projectID, nil, 10); !errors.Is(err, repository.ErrProjectNotFound) {
		t.Fatalf("removed member activity access: %v", err)
	}
}
