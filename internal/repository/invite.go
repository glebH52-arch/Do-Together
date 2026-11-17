package repository

import (
	"context"
	"do-together/internal/domain"
	"time"
	"uuid"
)

type InviteRepository interface {
	Create(ctx context.Context, invite *domain.Invite) error
	FindInviteeIDByEmail(ctx context.Context, email string) (uuid.UUID, error)
	List(ctx context.Context, inviteeID uuid.UUID) ([]*domain.Invite, error)
	CountPending(ctx context.Context, inviteeID uuid.UUID) (int, error)
	Accept(ctx context.Context, inviteeID, inviteID uuid.UUID, now time.Time) error
	Decline(ctx context.Context, inviteeID, inviteID uuid.UUID, now time.Time) error
}
