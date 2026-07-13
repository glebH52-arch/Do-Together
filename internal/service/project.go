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

	err = s.repository.Save(ctx, project)
	if err != nil {
		return nil, err
	}

	return project, nil
}
