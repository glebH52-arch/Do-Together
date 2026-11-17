package postgres

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/service"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
	"uuid"
)

func TestInviteByEmail(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	projectID, projectErr := testUUID("TEST_PROJECT_ID")
	actorID, actorErr := testUUID("TEST_ACTOR_ID")
	if url == "" || projectErr != nil || actorErr != nil {
		t.Skip("disposable PostgreSQL fixture is not configured")
	}
	ctx := context.Background()
	pool, err := NewPool(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	email := fmt.Sprintf("invite-%d@example.test", time.Now().UnixNano())
	var inviteeID uuid.UUID
	if err := pool.QueryRow(ctx, `INSERT INTO users(username,display_name,email,password_hash,status)
		VALUES ($1,$1,$2,'test','active') RETURNING id`, email, email).Scan(&inviteeID); err != nil {
		t.Fatal(err)
	}
	repo := NewPostgresInviteRepository(pool)
	inviteService := service.NewInviteService(repo)
	invite, err := inviteService.CreateByEmail(ctx, actorID, projectID, "  "+strings.ToUpper(email)+"  ", domain.ProjectMemberRoleMember)
	if err != nil || invite.InviteeID != inviteeID || invite.ProjectTitle == "" || invite.InviterName == "" {
		t.Fatalf("email invite: %+v %v", invite, err)
	}
	listed, err := repo.List(ctx, inviteeID)
	if err != nil || len(listed) != 1 || listed[0].ProjectTitle == "" || listed[0].InviterName == "" {
		t.Fatalf("invite list: %+v %v", listed, err)
	}
	count, err := repo.CountPending(ctx, inviteeID)
	if err != nil || count != 1 {
		t.Fatalf("pending invite count: %d %v", count, err)
	}
	var inviteNotifications int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM notifications WHERE invite_id = $1`, invite.ID).Scan(&inviteNotifications); err != nil || inviteNotifications != 0 {
		t.Fatalf("invite notification count: %d %v", inviteNotifications, err)
	}
	if _, err := inviteService.CreateByEmail(ctx, actorID, projectID, email, domain.ProjectMemberRoleMember); !errors.Is(err, domain.ErrPendingInviteAlreadyExists) {
		t.Fatalf("duplicate pending invite: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE invites SET expires_at = NOW() - INTERVAL '1 minute' WHERE id = $1`, invite.ID); err != nil {
		t.Fatal(err)
	}
	count, err = repo.CountPending(ctx, inviteeID)
	if err != nil || count != 0 {
		t.Fatalf("expired pending invite count: %d %v", count, err)
	}
	second, err := inviteService.Create(ctx, actorID, inviteeID, projectID, domain.ProjectMemberRoleMember)
	if err != nil || second.ID == invite.ID {
		t.Fatalf("invite after expiry: %+v %v", second, err)
	}
	listed, err = repo.List(ctx, inviteeID)
	if err != nil || len(listed) != 1 || listed[0].ID != second.ID {
		t.Fatalf("active pending invites: %+v %v", listed, err)
	}
	if err := repo.Decline(ctx, inviteeID, second.ID, time.Now()); err != nil {
		t.Fatal(err)
	}
	count, err = repo.CountPending(ctx, inviteeID)
	if err != nil || count != 0 {
		t.Fatalf("declined pending invite count: %d %v", count, err)
	}
	third, err := inviteService.Create(ctx, actorID, inviteeID, projectID, domain.ProjectMemberRoleMember)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Accept(ctx, inviteeID, third.ID, time.Now()); err != nil {
		t.Fatal(err)
	}
	count, err = repo.CountPending(ctx, inviteeID)
	if err != nil || count != 0 {
		t.Fatalf("accepted pending invite count: %d %v", count, err)
	}
}
