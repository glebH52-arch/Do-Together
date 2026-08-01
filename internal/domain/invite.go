package domain

import (
	"errors"
	"time"
)

type Invite struct {
	ID        int
	ProjectID int
	InviterID int
	InviteeID int
	Role      ProjectMemberRole
	Status    InviteStatus
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type InviteStatus string

const (
	InviteStatusPending  InviteStatus = "pending"
	InviteStatusAccepted InviteStatus = "accepted"
	InviteStatusDeclined InviteStatus = "declined"
	InviteStatusExpired  InviteStatus = "expired"
)

var (
	ErrInvalidInviterID           = errors.New("inviter id invalid")
	ErrInvalidInviteeID           = errors.New("invitee id invalid")
	ErrInvalidInviteRole          = errors.New("invite role invalid")
	ErrInviteNotPending           = errors.New("invite not pending")
	ErrInviteExpired              = errors.New("invite expired")
	ErrInviteNotFound             = errors.New("invite not found")
	ErrPendingInviteAlreadyExists = errors.New("pending invite already exists")
)

func NewInvite(inviterID, inviteeID, projectID int, role ProjectMemberRole) (*Invite, error) {
	if inviterID <= 0 {
		return nil, ErrInvalidInviterID
	}
	if inviteeID <= 0 {
		return nil, ErrInvalidInviteeID
	}
	if projectID <= 0 {
		return nil, ErrInvalidProjectID
	}
	err := role.validateInviteRole()
	if err != nil {
		return nil, err
	}

	return &Invite{
		ProjectID: projectID,
		InviterID: inviterID,
		InviteeID: inviteeID,
		Role:      role,
		Status:    InviteStatusPending,
		ExpiresAt: time.Now().Add(48 * time.Hour),
		CreatedAt: time.Now(),
	}, nil
}

func (i *Invite) Accept(now time.Time) error {

	if i.Status != InviteStatusPending {
		return ErrInviteNotPending
	}
	if now.After(i.ExpiresAt) {
		return ErrInviteExpired
	}

	i.Status = InviteStatusAccepted

	i.UpdatedAt = &now
	return nil
}

func (i *Invite) Decline(now time.Time) error {
	if i.Status != InviteStatusPending {
		return ErrInviteNotPending
	}
	if now.After(i.ExpiresAt) {
		return ErrInviteExpired
	}

	i.Status = InviteStatusDeclined

	i.UpdatedAt = &now
	return nil
}
