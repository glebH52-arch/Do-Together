package postgres

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/service"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestInviteByEmail(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	projectID, projectErr := strconv.Atoi(os.Getenv("TEST_PROJECT_ID"))
	actorID, actorErr := strconv.Atoi(os.Getenv("TEST_ACTOR_ID"))
	if url == "" || projectErr != nil || actorErr != nil || projectID <= 0 || actorID <= 0 {
		t.Skip("disposable PostgreSQL fixture is not configured")
	}
	ctx := context.Background()
	pool, err := NewPool(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	email := fmt.Sprintf("invite-%d@example.test", time.Now().UnixNano())
	var inviteeID int
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
}
