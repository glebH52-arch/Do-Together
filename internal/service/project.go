package service

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
)

type ProjectService struct {
	repository repository.ProjectRepository
}

func NewProjectService(repository repository.ProjectRepository) *ProjectService {
	return &ProjectService{
		repository: repository,
	}
}

func (s *ProjectService) CreateProject(ctx context.Context, userID int, title string, goal string) (*domain.Project, error) {
	project, err := domain.NewProject(userID, title, goal)
	if err != nil {
		return nil, err
	}

	err = s.repository.Create(ctx, userID, project)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) ListProjects(ctx context.Context, userID int) ([]*domain.Project, error) {
	projects, err := s.repository.List(ctx, userID)

	if err != nil {
		return nil, err
	}
	return projects, nil
}
func (s *ProjectService) UpdateProject(ctx context.Context, userID, id int, title *string, goal *string) (*domain.Project, error) {
	project, err := s.repository.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	err = project.Update(title, goal)
	if err != nil {
		return nil, err
	}
	err = s.repository.Update(ctx, userID, project)
	if err != nil {
		return nil, err
	}
	return project, nil

}
func (p *ProjectService) GetProject(ctx context.Context, userID, id int) (*domain.Project, error) {
	project, err := p.repository.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return project, nil
}
