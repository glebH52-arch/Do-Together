package domain

import (
	"errors"
	"time"
	"uuid"
)

type Invite struct {
	ID           uuid.UUID
	ProjectID    uuid.UUID
	InviterID    uuid.UUID
	InviteeID    uuid.UUID
	Role         ProjectMemberRole
	Status       InviteStatus
	ExpiresAt    time.Time
	CreatedAt    time.Time
	UpdatedAt    *time.Time
	ProjectTitle string
	InviterName  string
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

func NewInvite(inviterID, inviteeID, projectID uuid.UUID, role ProjectMemberRole) (*Invite, error) {
	if inviterID == uuid.Nil() {
		return nil, ErrInvalidInviterID
	}
	if inviteeID == uuid.Nil() {
		return nil, ErrInvalidInviteeID
	}
	if projectID == uuid.Nil() {
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
