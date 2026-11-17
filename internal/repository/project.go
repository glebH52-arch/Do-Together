package repository

import (
	"context"
	"do-together/internal/domain"
	"uuid"

	"errors"
	"sync"
)

var (
	ErrProjectNotFound     = errors.New("project not found")
	ErrNilProject          = errors.New("project is nil")
	ErrProjectAlreadySaved = errors.New("project already saved")
)

type ProjectRepository interface {
	Create(ctx context.Context, userID uuid.UUID, project *domain.Project) error
	GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Project, error)
	List(ctx context.Context, userID uuid.UUID, options ProjectListOptions) ([]*domain.Project, error)
	Update(ctx context.Context, userID uuid.UUID, project *domain.Project) error
}

type ProjectListOptions struct {
	Limit  int
	Offset int
	Status *domain.ProjectStatus
}

type MemoryProjectRepository struct {
	mu       sync.RWMutex
	projects map[uuid.UUID]*domain.Project
}

func NewMemoryProjectRepository() *MemoryProjectRepository {
	return &MemoryProjectRepository{
		mu:       sync.RWMutex{},
		projects: make(map[uuid.UUID]*domain.Project),
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

func (m *MemoryProjectRepository) Create(ctx context.Context, userID uuid.UUID, project *domain.Project) error {
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

	if project.ID != uuid.Nil() {
		return ErrProjectAlreadySaved
	}
	project.CreatedBy = userID
	project.ID = uuid.New()
	copyProject := cloneProject(project)
	m.projects[project.ID] = copyProject

	return nil
}

func (m *MemoryProjectRepository) GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Project, error) {

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
	if project.CreatedBy != userID {
		return nil, ErrProjectNotFound
	}

	copyProject := cloneProject(project)

	return copyProject, nil

}

func (m *MemoryProjectRepository) List(ctx context.Context, userID uuid.UUID, options ProjectListOptions) ([]*domain.Project, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	listProjects := make([]*domain.Project, 0, len(m.projects))

	for _, project := range m.projects {
		if project.CreatedBy == userID {
			clonedProject := cloneProject(project)
			listProjects = append(listProjects, clonedProject)
		}
	}

	return listProjects, nil
}
func (m *MemoryProjectRepository) Update(ctx context.Context, userID uuid.UUID, project *domain.Project) error {

	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}

	if project == nil {
		return ErrNilProject
	}

	storedProject, found := m.projects[project.ID]

	if !found {
		return ErrProjectNotFound
	}
	if storedProject.CreatedBy != userID {
		return ErrProjectNotFound
	}

	copyProject := cloneProject(project)
	copyProject.CreatedBy = storedProject.CreatedBy
	m.projects[project.ID] = copyProject

	return nil

}
