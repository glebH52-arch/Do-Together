package service

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"errors"
	"testing"
	"time"
	"uuid"
)

type activityRepositoryStub struct {
	items  []*domain.ProjectActivity
	cursor *repository.ActivityCursor
	limit  int
}

func (r *activityRepositoryStub) List(_ context.Context, _, _ uuid.UUID, cursor *repository.ActivityCursor, limit int) ([]*domain.ProjectActivity, error) {
	r.cursor, r.limit = cursor, limit
	return r.items, nil
}

func TestActivityCursorRoundTrip(t *testing.T) {
	want := repository.ActivityCursor{OccurredAt: time.Date(2026, 11, 17, 13, 21, 45, 123456789, time.UTC), ID: uuid.New()}
	encoded, err := EncodeActivityCursor(want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeActivityCursor(encoded)
	if err != nil || got.ID != want.ID || !got.OccurredAt.Equal(want.OccurredAt) {
		t.Fatalf("round trip: %#v %v", got, err)
	}
	for _, invalid := range []string{"!", "e30", "eyJvY2N1cnJlZF9hdCI6ImJhZCIsImlkIjoiMDAwMDAwMDAtMDAwMC0wMDAwLTAwMDAtMDAwMDAwMDAwMDAwIn0"} {
		if _, err := DecodeActivityCursor(invalid); !errors.Is(err, ErrInvalidActivityCursor) {
			t.Fatalf("%q: %v", invalid, err)
		}
	}
}

func TestProjectActivityPage(t *testing.T) {
	now := time.Now()
	repo := &activityRepositoryStub{items: []*domain.ProjectActivity{
		{ID: uuid.New(), OccurredAt: now},
		{ID: uuid.New(), OccurredAt: now.Add(-time.Second)},
		{ID: uuid.New(), OccurredAt: now.Add(-2 * time.Second)},
	}}
	service := NewProjectActivityService(repo)
	page, err := service.List(context.Background(), uuid.New(), uuid.New(), "", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 2 || page.NextCursor == nil || repo.limit != 3 {
		t.Fatalf("page: %+v repository limit %d", page, repo.limit)
	}
	cursor, err := DecodeActivityCursor(*page.NextCursor)
	if err != nil || cursor.ID != page.Items[1].ID {
		t.Fatalf("next cursor: %+v %v", cursor, err)
	}
}
