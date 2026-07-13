package repository

import (
	"context"
	"do-together/internal/domain"
	"errors"
	"sync"
)

var (
	ErrProjectNotFound = errors.New("project not found")
	ErrContextDone     = errors.New("contexct is unepexted done")
)

type ProjectRepository interface {
	Save(ctx context.Context, project *domain.Project) error
	GetByID(ctx context.Context, id int) (*domain.Project, error)
}

type MemoryProjectRepository struct {
	mu       sync.RWMutex
	nextID   int
	projects map[int]*domain.Project
}

func NewMemoryProjectRepository() *MemoryProjectRepository {
	return &MemoryProjectRepository{
		mu:       sync.RWMutex{},
		nextID:   1,
		projects: make(map[int]*domain.Project),
	}
}

func (m *MemoryProjectRepository) Save(ctx context.Context, project *domain.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}

	project.ID = m.nextID
	m.projects[m.nextID] = project
	m.nextID++

	return nil
}

func (m *MemoryProjectRepository) GetByID(ctx context.Context, id int) (*domain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	project, ok := m.projects[id]

	if !ok {
		return nil, ErrProjectNotFound
	}

	return project, nil

}
