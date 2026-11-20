package service

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"uuid"
)

type TaskCommentService struct {
	repository repository.TaskCommentRepository
}

func NewTaskCommentService(repository repository.TaskCommentRepository) *TaskCommentService {
	return &TaskCommentService{repository: repository}
}

func (s *TaskCommentService) Create(ctx context.Context, actorID, projectID, taskID uuid.UUID, body string) (*domain.TaskComment, error) {
	body, err := domain.ValidateCommentBody(body)
	if err != nil {
		return nil, err
	}
	return s.repository.Create(ctx, actorID, projectID, taskID, body)
}

func (s *TaskCommentService) List(ctx context.Context, actorID, projectID, taskID uuid.UUID, limit, offset int) ([]*domain.TaskComment, error) {
	if limit <= 0 || limit > 100 || offset < 0 {
		return nil, domain.ErrInvalidCommentPage
	}
	return s.repository.List(ctx, actorID, projectID, taskID, limit, offset)
}

func (s *TaskCommentService) Update(ctx context.Context, actorID, projectID, taskID uuid.UUID, commentID uuid.UUID, body string) error {
	body, err := domain.ValidateCommentBody(body)
	if err != nil {
		return err
	}
	return s.repository.Update(ctx, actorID, projectID, taskID, commentID, body)
}

func (s *TaskCommentService) Delete(ctx context.Context, actorID, projectID, taskID uuid.UUID, commentID uuid.UUID) error {
	return s.repository.Delete(ctx, actorID, projectID, taskID, commentID)
}
