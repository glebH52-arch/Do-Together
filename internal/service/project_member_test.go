package service

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"errors"
	"testing"
	"uuid"
)

type projectMemberRepositoryStub struct {
	members []*domain.ProjectMember
}

func (s *projectMemberRepositoryStub) List(context.Context, uuid.UUID, uuid.UUID) ([]*domain.ProjectMember, error) {
	return s.members, nil
}
func (s *projectMemberRepositoryStub) Add(context.Context, uuid.UUID, *domain.ProjectMember) error {
	return nil
}
func (s *projectMemberRepositoryStub) UpdateRole(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, domain.ProjectMemberRole) error {
	return nil
}
func (s *projectMemberRepositoryStub) Remove(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error {
	return nil
}

func TestProjectMemberRolePolicy(t *testing.T) {
	projectID := uuid.New()
	creatorID := uuid.New()
	adminID := uuid.New()
	memberID := uuid.New()
	repo := &projectMemberRepositoryStub{members: []*domain.ProjectMember{
		{ProjectID: projectID, UserID: creatorID, Role: domain.ProjectMemberRoleCreator},
		{ProjectID: projectID, UserID: adminID, Role: domain.ProjectMemberRoleAdmin},
		{ProjectID: projectID, UserID: memberID, Role: domain.ProjectMemberRoleMember},
	}}
	service := NewProjectMemberService(repo)

	tests := []struct {
		name   string
		actor  uuid.UUID
		target uuid.UUID
		role   domain.ProjectMemberRole
		allow  bool
	}{
		{"creator promotes member", creatorID, memberID, domain.ProjectMemberRoleAdmin, true},
		{"creator demotes admin", creatorID, adminID, domain.ProjectMemberRoleMember, true},
		{"creator cannot change creator", creatorID, creatorID, domain.ProjectMemberRoleMember, false},
		{"admin cannot promote member", adminID, memberID, domain.ProjectMemberRoleAdmin, false},
		{"admin may keep member role", adminID, memberID, domain.ProjectMemberRoleMember, true},
		{"admin cannot change admin", adminID, adminID, domain.ProjectMemberRoleMember, false},
		{"member cannot change role", memberID, adminID, domain.ProjectMemberRoleMember, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := service.UpdateRole(context.Background(), test.actor, projectID, test.target, test.role)
			if test.allow && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !test.allow && !errors.Is(err, repository.ErrForbidden) {
				t.Fatalf("expected forbidden, got %v", err)
			}
		})
	}
}

func TestProjectMemberRemovalPolicy(t *testing.T) {
	projectID := uuid.New()
	creatorID := uuid.New()
	adminID := uuid.New()
	memberID := uuid.New()
	repo := &projectMemberRepositoryStub{members: []*domain.ProjectMember{
		{ProjectID: projectID, UserID: creatorID, Role: domain.ProjectMemberRoleCreator},
		{ProjectID: projectID, UserID: adminID, Role: domain.ProjectMemberRoleAdmin},
		{ProjectID: projectID, UserID: memberID, Role: domain.ProjectMemberRoleMember},
	}}
	service := NewProjectMemberService(repo)

	if err := service.Remove(context.Background(), adminID, projectID, memberID); err != nil {
		t.Fatalf("admin removing member: %v", err)
	}
	if err := service.Remove(context.Background(), adminID, projectID, adminID); !errors.Is(err, repository.ErrForbidden) {
		t.Fatalf("admin removing admin: %v", err)
	}
	if err := service.Remove(context.Background(), creatorID, projectID, adminID); err != nil {
		t.Fatalf("creator removing admin: %v", err)
	}
	if err := service.Remove(context.Background(), creatorID, projectID, creatorID); !errors.Is(err, repository.ErrForbidden) {
		t.Fatalf("creator removing creator: %v", err)
	}
}

func TestAdminCannotAddAdmin(t *testing.T) {
	projectID := uuid.New()
	adminID := uuid.New()
	repo := &projectMemberRepositoryStub{members: []*domain.ProjectMember{
		{ProjectID: projectID, UserID: adminID, Role: domain.ProjectMemberRoleAdmin},
	}}
	service := NewProjectMemberService(repo)
	if err := service.Add(context.Background(), adminID, projectID, uuid.New(), domain.ProjectMemberRoleAdmin); !errors.Is(err, repository.ErrForbidden) {
		t.Fatalf("admin adding admin: %v", err)
	}
	if err := service.Add(context.Background(), adminID, projectID, uuid.New(), domain.ProjectMemberRoleMember); err != nil {
		t.Fatalf("admin adding member: %v", err)
	}
}
