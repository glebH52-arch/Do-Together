package repository

import (
	"context"
	"do-together/internal/domain"
	"errors"
	"testing"
)

func TestMemoryProjectRepository_SaveAndGetByID(t *testing.T) {
	repository := NewMemoryProjectRepository()
	project, err := domain.NewProject(123, "learn go ", "jind job in 3 months")

	if err != nil {
		t.Fatalf("Save returned unexpected error: %v", err)
	}

	err = repository.Save(context.Background(), project)

	if err != nil {
		t.Fatalf("Save returned unexpected error: %v", err)
	}

	if project.ID == 0 {
		t.Errorf("expected project ID %v, got %v", 1, project.ID)
	}

	projectget, err := repository.GetByID(context.Background(), 1)

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
	projectget, err := repository.GetByID(context.Background(), 1)

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
