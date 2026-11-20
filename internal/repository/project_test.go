package repository

import (
	"context"
	"do-together/internal/domain"
	"errors"
	"testing"
	"uuid"
)

func TestMemoryProjectRepository_SaveAndGetByID(t *testing.T) {
	repository := NewMemoryProjectRepository()
	userID := uuid.New()
	project, err := domain.NewProject(userID, "learn go ", "jind job in 3 months")

	if err != nil {
		t.Fatalf("Save returned unexpected error: %v", err)
	}

	err = repository.Create(context.Background(), userID, project)

	if err != nil {
		t.Fatalf("Save returned unexpected error: %v", err)
	}

	if project.ID == uuid.Nil() {
		t.Error("expected generated project ID")
	}

	projectget, err := repository.GetByID(context.Background(), userID, project.ID)

	if err != nil {
		t.Fatalf("GetByID returned unexpected error: %v", err)
	}

	if projectget.Goal != project.Goal {
		t.Errorf("expected project Goal %v, got %v", project.Goal, projectget.Goal)
	}

	if projectget.Title != project.Title {
		t.Errorf("expected project Title %v, got %v", project.Title, projectget.Title)
	}
}

func TestMemoryProjectRepository_GetByID_NotFound(t *testing.T) {
	repository := NewMemoryProjectRepository()
	projectget, err := repository.GetByID(context.Background(), uuid.New(), uuid.New())

	if err == nil {
		t.Fatal("expected ErrProjectNotFound, got nil")
	}

	if projectget != nil {
		t.Fatal("expected nil project on foud error")
	}

	// Пришла не та ошибка
	if !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("expected  ErrProjectNotFound, got %v", err)
	}

}

func TestMemoryProjectRepository_StoredProjectIsIndependent(t *testing.T) {
	repository := NewMemoryProjectRepository()
	userID := uuid.New()
	project, err := domain.NewProject(userID, "learn go ", "jind job in 3 months")

	if err != nil {
		t.Fatalf("NewProject returned unexpected error: %v", err)
	}

	originalTitle := project.Title

	err = repository.Create(context.Background(), userID, project)

	if err != nil {
		t.Fatalf("Save returned unexpected error: %v", err)
	}

	project.Title = "changed"
	gotProject, err := repository.GetByID(context.Background(), userID, project.ID)

	if gotProject == nil {
		t.Fatal("GetByID returned nil project without error")
	}

	if err != nil {
		t.Fatalf("GetByID returned unexpected error: %v", err)
	}

	if gotProject.Title != originalTitle {
		t.Errorf("expected project Title %v, got %v", originalTitle, gotProject.Title)
	}

	gotProject.Title = "Hack"

	gotAgain, err := repository.GetByID(context.Background(), userID, project.ID)

	if gotAgain == nil {
		t.Fatal("GetByID returned nil project without error")
	}

	if err != nil {
		t.Fatal("expected ErrProjectNotFound, got nil")
	}

	if gotAgain.Title != originalTitle {
		t.Errorf("expected project Title %v, got %v", originalTitle, gotAgain.Title)
	}
}
