package repository

import (
	"context"
	"do-together/internal/domain"
	"errors"
	"sort"
	"sync"
)

var (
	ErrProjectNotFound     = errors.New("project not found")
	ErrNilProject          = errors.New("project is nil")
	ErrProjectAlreadySaved = errors.New("project already saved")
)

type ProjectRepository interface {
	Create(ctx context.Context, userID int, project *domain.Project) error
	GetByID(ctx context.Context, userID, id int) (*domain.Project, error)
	List(ctx context.Context, userID int) ([]*domain.Project, error)
	Update(ctx context.Context, userID int, project *domain.Project) error
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

func (m *MemoryProjectRepository) Create(ctx context.Context, userID int, project *domain.Project) error {
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
	project.CreatedBy = userID
	project.ID = m.nextID
	copyProject := cloneProject(project)
	m.projects[m.nextID] = copyProject
	m.nextID++

	return nil
}

func (m *MemoryProjectRepository) GetByID(ctx context.Context, userID, id int) (*domain.Project, error) {

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

func (m *MemoryProjectRepository) List(ctx context.Context, userID int) ([]*domain.Project, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	ids := make([]int, 0, len(m.projects))
	for id := range m.projects {

		ids = append(ids, id)
	}

	sort.Ints(ids)
	listProjects := make([]*domain.Project, 0, len(ids))

	for _, id := range ids {
		project := m.projects[id]
		if project.CreatedBy == userID {
			clonedProject := cloneProject(project)
			listProjects = append(listProjects, clonedProject)
		}
	}

	return listProjects, nil
}
func (m *MemoryProjectRepository) Update(ctx context.Context, userID int, project *domain.Project) error {

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
