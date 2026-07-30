package service

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
)

type ProjectMemberService struct {
	repository repository.ProjectMemberRepository
}

func NewProjectMemberService(repository repository.ProjectMemberRepository) *ProjectMemberService {
	return &ProjectMemberService{
		repository: repository,
	}
}

func (p *ProjectMemberService) List(ctx context.Context, actorID, projectID int) ([]*domain.ProjectMember, error) {
	if actorID <= 0 {
		return nil, repository.ErrProjectNotFound
	}
	if projectID <= 0 {
		return nil, repository.ErrProjectNotFound
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	projectMembrs, err := p.repository.List(ctx, actorID, projectID)
	if err != nil {
		return nil, err
	}
	return projectMembrs, nil
}
func (p *ProjectMemberService) Add(ctx context.Context, actorID int, projectID int, userID int, role domain.ProjectMemberRole) error {
	if actorID <= 0 {
		return repository.ErrProjectNotFound
	}
	if projectID <= 0 {
		return repository.ErrProjectNotFound
	}
	if userID <= 0 {
		return repository.ErrUserNotFound
	}
	err := role.IsValid()
	if err != nil {
		return err
	}
	if role == domain.ProjectMemberRoleCreator {
		return repository.ErrForbidden
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	_, err = p.authorizeActor(ctx, actorID, projectID)
	if err != nil {
		return err
	}

	member, err := domain.NewProjectMember(projectID, userID, role)
	if err != nil {
		return err
	}
	err = p.repository.Add(ctx, actorID, member)
	if err != nil {
		return err
	}
	return nil
}
func (p *ProjectMemberService) UpdateRole(ctx context.Context, actorID, projectID, targetUserID int, role domain.ProjectMemberRole) error {
	if actorID <= 0 {
		return repository.ErrProjectNotFound
	}
	if projectID <= 0 {
		return repository.ErrProjectNotFound
	}
	if targetUserID <= 0 {
		return repository.ErrUserNotFound
	}
	err := role.IsValid()
	if err != nil {
		return err
	}
	if role == domain.ProjectMemberRoleCreator {
		return repository.ErrForbidden
	}
	projectMembers, err := p.authorizeActor(ctx, actorID, projectID)
	if err != nil {
		return err
	}
	err = p.authorizeTarget(ctx, projectMembers, targetUserID)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	err = p.repository.UpdateRole(ctx, actorID, projectID, targetUserID, role)
	if err != nil {
		return err
	}
	return nil
}
func (p *ProjectMemberService) Remove(ctx context.Context, actorID, projectID, targetUserID int) error {
	if actorID <= 0 {
		return repository.ErrProjectNotFound
	}
	if projectID <= 0 {
		return repository.ErrProjectNotFound
	}
	if targetUserID <= 0 {
		return repository.ErrUserNotFound
	}
	projectMembers, err := p.authorizeActor(ctx, actorID, projectID)
	if err != nil {
		return err
	}
	err = p.authorizeTarget(ctx, projectMembers, targetUserID)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	err = p.repository.Remove(ctx, actorID, projectID, targetUserID)
	if err != nil {
		return err
	}
	return nil
}

func (p *ProjectMemberService) authorizeActor(ctx context.Context, actorID int, projectID int) ([]*domain.ProjectMember, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	projectMembers, err := p.repository.List(ctx, actorID, projectID)
	if err != nil {
		return nil, err
	}
	for _, member := range projectMembers {
		if member.UserID != actorID {
			continue
		}

		if member.Role != domain.ProjectMemberRoleCreator &&
			member.Role != domain.ProjectMemberRoleAdmin {
			return nil, repository.ErrForbidden
		}

		return projectMembers, nil
	}

	return nil, repository.ErrProjectNotFound

}
func (p *ProjectMemberService) authorizeTarget(ctx context.Context, projectMembers []*domain.ProjectMember, targetUserID int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	for _, member := range projectMembers {
		if member.UserID != targetUserID {
			continue
		}

		if member.Role == domain.ProjectMemberRoleCreator {
			return repository.ErrForbidden
		}

		return nil
	}

	return repository.ErrProjectNotFound

}
