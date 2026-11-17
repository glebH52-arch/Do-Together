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

type taskSearchRepository struct {
	searchCalled bool
}

func (r *taskSearchRepository) Create(context.Context, *domain.Task) (*domain.Task, error) {
	panic("unexpected call")
}
func (r *taskSearchRepository) List(context.Context, uuid.UUID, uuid.UUID) ([]*domain.Task, error) {
	panic("unexpected call")
}
func (r *taskSearchRepository) Update(context.Context, *domain.Task, *domain.Task, uuid.UUID, uuid.UUID) error {
	panic("unexpected call")
}
func (r *taskSearchRepository) Remove(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error {
	panic("unexpected call")
}
func (r *taskSearchRepository) GetByID(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (*domain.Task, error) {
	panic("unexpected call")
}
func (r *taskSearchRepository) Search(context.Context, uuid.UUID, uuid.UUID, repository.TaskSearchOptions) ([]*domain.Task, error) {
	r.searchCalled = true
	return []*domain.Task{}, nil
}
func (r *taskSearchRepository) History(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, int, int) ([]*domain.TaskHistory, error) {
	panic("unexpected call")
}

func TestSearchTasksDueRange(t *testing.T) {
	actorID := uuid.New()
	projectID := uuid.New()
	from := time.Date(2026, time.November, 17, 0, 0, 0, 0, time.UTC)
	to := from.Add(24 * time.Hour)

	tests := []struct {
		name    string
		after   time.Time
		before  time.Time
		wantErr bool
	}{
		{name: "valid half-open range", after: from, before: to},
		{name: "equal bounds", after: from, before: from, wantErr: true},
		{name: "reversed bounds", after: to, before: from, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &taskSearchRepository{}
			service := NewTaskService(repo)
			_, err := service.SearchTasks(context.Background(), actorID, projectID, repository.TaskSearchOptions{
				DueAfter:  &tt.after,
				DueBefore: &tt.before,
				Limit:     20,
			})
			if tt.wantErr {
				if !errors.Is(err, repository.ErrInvalidTaskSearch) {
					t.Fatalf("expected invalid search error, got %v", err)
				}
				if repo.searchCalled {
					t.Fatal("repository search was called for an invalid range")
				}
				return
			}
			if err != nil {
				t.Fatalf("valid range: %v", err)
			}
			if !repo.searchCalled {
				t.Fatal("repository search was not called for a valid range")
			}
		})
	}
}
