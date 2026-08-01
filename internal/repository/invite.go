package repository

import (
	"context"
	"do-together/internal/domain"
	"time"
)

type InviteRepository interface {
	Create(ctx context.Context, invite *domain.Invite) error
	List(ctx context.Context, inviteeID int) ([]*domain.Invite, error)
	Accept(ctx context.Context, inviteeID, inviteID int, now time.Time) error
	Decline(ctx context.Context, inviteeID, inviteID int, now time.Time) error
}
