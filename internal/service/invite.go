package service

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"time"
)

type InviteService struct {
	repository repository.InviteRepository
}

func NewInviteService(repository repository.InviteRepository) *InviteService {
	return &InviteService{
		repository: repository,
	}
}

func (s *InviteService) Create(ctx context.Context, inviterID, inviteeID, projectID int, role domain.ProjectMemberRole) (*domain.Invite, error) {

	if inviteeID <= 0 {
		return nil, repository.ErrUserNotFound
	}
	if inviterID <= 0 {
		return nil, repository.ErrUserNotFound
	}
	if projectID <= 0 {
		return nil, repository.ErrProjectNotFound
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	invite, err := domain.NewInvite(inviterID, inviteeID, projectID, role)
	if err != nil {
		return nil, err
	}
	err = s.repository.Create(ctx, invite)
	if err != nil {
		return nil, err
	}
	return invite, nil
}
func (s *InviteService) Accept(ctx context.Context, inviteeID, inviteID int) error {
	if inviteeID <= 0 {
		return repository.ErrUserNotFound
	}
	if inviteID <= 0 {
		return domain.ErrInviteNotFound
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return s.repository.Accept(ctx, inviteeID, inviteID, time.Now())

}
func (s *InviteService) Decline(ctx context.Context, inviteeID, inviteID int) error {
	if inviteeID <= 0 {
		return repository.ErrUserNotFound
	}
	if inviteID <= 0 {
		return domain.ErrInviteNotFound
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return s.repository.Decline(ctx, inviteeID, inviteID, time.Now())

}
func (s *InviteService) List(ctx context.Context, inviteeID int) ([]*domain.Invite, error) {
	if inviteeID <= 0 {
		return nil, repository.ErrUserNotFound
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	invites, err := s.repository.List(ctx, inviteeID)
	if err != nil {
		return nil, err
	}
	return invites, nil
}
