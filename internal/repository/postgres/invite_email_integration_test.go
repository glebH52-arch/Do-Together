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
	if err := pool.QueryRow(ctx, `INSERT INTO users(username,email,password_hash) VALUES ($1,$2,'test') RETURNING id`, email, email).Scan(&inviteeID); err != nil {
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
	items, err := NewNotificationRepository(pool).List(ctx, inviteeID, 20, 0)
	if err != nil || len(items) != 1 || items[0].Kind != "invitation" || items[0].InviteID == nil || *items[0].InviteID != invite.ID {
		t.Fatalf("invite notification: %+v %v", items, err)
	}
	if _, err := inviteService.CreateByEmail(ctx, actorID, projectID, email, domain.ProjectMemberRoleMember); !errors.Is(err, domain.ErrPendingInviteAlreadyExists) {
		t.Fatalf("duplicate pending invite: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE invites SET expires_at = NOW() - INTERVAL '1 minute' WHERE id = $1`, invite.ID); err != nil {
		t.Fatal(err)
	}
	second, err := inviteService.Create(ctx, actorID, inviteeID, projectID, domain.ProjectMemberRoleMember)
	if err != nil || second.ID == invite.ID {
		t.Fatalf("invite after expiry: %+v %v", second, err)
	}
	listed, err = repo.List(ctx, inviteeID)
	if err != nil || len(listed) != 2 || listed[0].Status != domain.InviteStatusExpired {
		t.Fatalf("expired invite: %+v %v", listed, err)
	}
	rollbackEmail := fmt.Sprintf("invite-rollback-%d@example.test", time.Now().UnixNano())
	var rollbackInviteeID uuid.UUID
	if err := pool.QueryRow(ctx, `INSERT INTO users(username,email,password_hash) VALUES ($1,$2,'test') RETURNING id`, rollbackEmail, rollbackEmail).Scan(&rollbackInviteeID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `CREATE FUNCTION test_reject_invite_notification() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'forced notification failure'; END $$`); err != nil {
		t.Fatal(err)
	}
	defer pool.Exec(ctx, `DROP FUNCTION IF EXISTS test_reject_invite_notification() CASCADE`)
	if _, err := pool.Exec(ctx, `CREATE TRIGGER test_reject_invite_notification BEFORE INSERT ON notifications FOR EACH ROW EXECUTE FUNCTION test_reject_invite_notification()`); err != nil {
		t.Fatal(err)
	}
	rollbackInvite, err := domain.NewInvite(actorID, rollbackInviteeID, projectID, domain.ProjectMemberRoleMember)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, rollbackInvite); err == nil {
		t.Fatal("notification failure did not roll back invite creation")
	}
	var inviteCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM invites WHERE project_id = $1 AND invitee_id = $2`, projectID, rollbackInviteeID).Scan(&inviteCount); err != nil {
		t.Fatal(err)
	}
	if inviteCount != 0 {
		t.Fatalf("invite remains after notification rollback: %d", inviteCount)
	}
}
