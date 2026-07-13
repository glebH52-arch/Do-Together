package repository

import (
	"context"
	"do-together/internal/domain"
	"errors"
	"sync"
)

var (
	ErrProjectNotFound     = errors.New("project not found")
	ErrNilProject          = errors.New("project is nil")
	ErrProjectAlreadySaved = errors.New("project already saved")
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

func cloneProject(project *domain.Project) *domain.Project {
	if project == nil {
		return nil
	}

	cloned := *project
	if project.UpdatedAt != nil {
		updatedAtCopy := *project.UpdatedAt
		cloned.UpdatedAt = &updatedAtCopy
	}

	return &cloned
}

func (m *MemoryProjectRepository) Save(ctx context.Context, project *domain.Project) error {
	if project == nil {
		return ErrNilProject
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}

	if project.ID != 0 {
		return ErrProjectAlreadySaved
	}
	project.ID = m.nextID
	copyProject := cloneProject(project)
	m.projects[m.nextID] = copyProject
	m.nextID++

	return nil
}

func (m *MemoryProjectRepository) GetByID(ctx context.Context, id int) (*domain.Project, error) {

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	project, ok := m.projects[id]

	if !ok {
		return nil, ErrProjectNotFound
	}

	copyProject := cloneProject(project)

	return copyProject, nil

}
