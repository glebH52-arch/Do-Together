package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestNewProject_Success(t *testing.T) {
	// Arrange: подготовить userID, title и goal

	// Act: вызвать NewProject

	// Assert: проверить ошибку и поля результата

	userID := 123456
	title := "  get a job  "
	goal := "  get a job as a junior Go developer  "

	expectedTitle := "get a job"
	expectedGoal := "get a job as a junior Go developer"

	project, err := NewProject(userID, title, goal)

	if err != nil {
		t.Fatalf("NewProject() returned unexpected error: %v", err)
	}
	if project == nil {
		t.Fatal("NewProject() returned nil project without error")
	}
	if project.CreatedBy != userID {
		t.Errorf("expected project userID %v, got %v", userID, project.CreatedBy)
	}

	if project.Goal != expectedGoal {
		t.Errorf("expected project goal %v, got %v", expectedGoal, project.Goal)
	}

	if project.Status != ProjectStatusActive {
		t.Errorf("expected project Status %v, got %v", ProjectStatusActive, project.Status)
	}

	if project.Title != expectedTitle {
		t.Errorf("expected project title %v, got %v", expectedTitle, project.Title)
	}

	if project.ID != 0 {
		t.Errorf("expected project ID %v, got %v", 0, project.ID)

	}

	if project.UpdatedAt != nil {
		t.Errorf("expected project UpdatedAt %v, got %v", nil, project.UpdatedAt)
	}

	if project.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

}

func TestNewProject_EmptyTitle(t *testing.T) {
	// Arrange
	userID := 123456
	title := "   "
	goal := "get a job"

	// Act
	project, err := NewProject(userID, title, goal)

	// Assert
	// 1. project должен быть nil
	// 2. err должен определяться как ErrTitleEmpty через errors.Is

	if err == nil {
		t.Fatal("expected ErrTitleEmpty, got nil")
	}

	if project != nil {
		t.Fatal("expected nil project on validation error")
	}

	// Пришла не та ошибка
	if !errors.Is(err, ErrTitleEmpty) {
		t.Fatalf("expected ErrTitleEmpty, got %v", err)
	}

}
func TestNewProject_TitleTooLong(t *testing.T) {
	// Arrange
	userID := 123456
	title := strings.Repeat("я", 101)
	goal := "get a job"

	// Act
	project, err := NewProject(userID, title, goal)

	// Assert
	if err == nil {
		t.Fatal("expected ErrTitleTooLong, got nil")
	}

	if project != nil {
		t.Fatal("expected nil project on validation error")
	}

	// Пришла не та ошибка
	if !errors.Is(err, ErrTitleTooLong) {
		t.Fatalf("expected ErrTitleTooLong, got %v", err)
	}
}

func TestNewProject_EmptyGoal(t *testing.T) {
	// Arrange
	userID := 123456
	title := "12343"
	goal := "		"

	// Act
	project, err := NewProject(userID, title, goal)

	// Assert
	// 1. project должен быть nil
	// 2. err должен определяться как ErrTitleEmpty через errors.Is

	if err == nil {
		t.Fatal("expected ErrGoalEmpty, got nil")
	}

	if project != nil {
		t.Fatal("expected nil project on validation error")
	}

	// Пришла не та ошибка
	if !errors.Is(err, ErrGoalEmpty) {
		t.Fatalf("expected ErrGoalEmpty, got %v", err)
	}

}
func TestNewProject_GoalTooLong(t *testing.T) {
	// Arrange
	userID := 123456
	title := "get a job"
	goal := strings.Repeat("я", 201)

	// Act
	project, err := NewProject(userID, title, goal)

	// Assert
	if err == nil {
		t.Fatal("expected ErrGoalTooLong, got nil")
	}

	if project != nil {
		t.Fatal("expected nil project on validation error")
	}

	// Пришла не та ошибка
	if !errors.Is(err, ErrGoalTooLong) {
		t.Fatalf("expected ErrGoalTooLong, got %v", err)
	}
}

func TestNewProject_TitleTooAtMaxLenght(t *testing.T) {
	// Arrange
	userID := 123456
	title := strings.Repeat("я", 100)
	goal := "get a job"

	// Act
	project, err := NewProject(userID, title, goal)

	// Assert
	if err != nil {
		t.Fatalf("NewProject() returned unexpected error: %v", err)
	}
	if project == nil {
		t.Fatal("NewProject() returned nil project without error")
	}

}
