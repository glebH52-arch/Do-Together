package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
	"uuid"
)

func TestNewProject_Success(t *testing.T) {
	userID := uuid.New()
	project, err := NewProject(userID, "  get a job  ", "  get a job as a junior Go developer  ")
	if err != nil {
		t.Fatalf("NewProject() returned unexpected error: %v", err)
	}
	if project == nil {
		t.Fatal("NewProject() returned nil project without error")
	}
	if project.CreatedBy != userID {
		t.Errorf("expected project userID %v, got %v", userID, project.CreatedBy)
	}
	if project.Goal != "get a job as a junior Go developer" {
		t.Errorf("unexpected project goal %q", project.Goal)
	}
	if project.Status != ProjectStatusActive {
		t.Errorf("expected project status %v, got %v", ProjectStatusActive, project.Status)
	}
	if project.Title != "get a job" {
		t.Errorf("unexpected project title %q", project.Title)
	}
	if project.ID != uuid.Nil() {
		t.Errorf("expected empty project ID, got %v", project.ID)
	}
	if project.UpdatedAt != nil {
		t.Errorf("expected nil UpdatedAt, got %v", project.UpdatedAt)
	}
	if project.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestNewProject_EmptyTitle(t *testing.T) {
	project, err := NewProject(uuid.New(), "   ", "get a job")
	if project != nil || !errors.Is(err, ErrTitleEmpty) {
		t.Fatalf("project=%v error=%v, want nil and ErrTitleEmpty", project, err)
	}
}

func TestNewProject_TitleTooLong(t *testing.T) {
	project, err := NewProject(uuid.New(), strings.Repeat("я", 101), "get a job")
	if project != nil || !errors.Is(err, ErrTitleTooLong) {
		t.Fatalf("project=%v error=%v, want nil and ErrTitleTooLong", project, err)
	}
}

func TestNewProject_EmptyGoal(t *testing.T) {
	project, err := NewProject(uuid.New(), "title", "\t\t")
	if project != nil || !errors.Is(err, ErrGoalEmpty) {
		t.Fatalf("project=%v error=%v, want nil and ErrGoalEmpty", project, err)
	}
}

func TestNewProject_GoalTooLong(t *testing.T) {
	project, err := NewProject(uuid.New(), "title", strings.Repeat("я", 201))
	if project != nil || !errors.Is(err, ErrGoalTooLong) {
		t.Fatalf("project=%v error=%v, want nil and ErrGoalTooLong", project, err)
	}
}

func TestNewProject_TitleAtMaxLength(t *testing.T) {
	project, err := NewProject(uuid.New(), strings.Repeat("я", 100), "get a job")
	if err != nil || project == nil {
		t.Fatalf("project=%v error=%v", project, err)
	}
}

func TestProjectRestore(t *testing.T) {
	archivedAt := time.Now().Add(-time.Hour)
	project := Project{
		Status:    ProjectStatusArchived,
		UpdatedAt: &archivedAt,
	}

	project.Restore()

	if project.Status != ProjectStatusActive {
		t.Fatalf("status = %q, want %q", project.Status, ProjectStatusActive)
	}
	if project.UpdatedAt == nil || !project.UpdatedAt.After(archivedAt) {
		t.Fatalf("updated_at = %v, want after %v", project.UpdatedAt, archivedAt)
	}

	restoredAt := project.UpdatedAt
	project.Restore()
	if project.UpdatedAt != restoredAt {
		t.Fatal("repeated restore changed updated_at")
	}
}

func TestProjectArchiveRestoreRoundTrip(t *testing.T) {
	project := Project{Status: ProjectStatusActive}
	project.Archive()
	if project.Status != ProjectStatusArchived {
		t.Fatalf("archive status = %q", project.Status)
	}
	project.Restore()
	if project.Status != ProjectStatusActive {
		t.Fatalf("restore status = %q", project.Status)
	}
}
