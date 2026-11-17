package domain

import (
	"encoding/json"
	"time"
	"uuid"
)

type Notification struct {
	ID          uuid.UUID
	RecipientID uuid.UUID
	ProjectID   uuid.UUID
	TaskID      *uuid.UUID
	InviteID    *uuid.UUID
	Kind        string
	Payload     json.RawMessage
	CreatedAt   time.Time
	ReadAt      *time.Time
}

type NotificationKind string

const (
	NotificationKindAssignmentChanged NotificationKind = "assignment_changed"
	NotificationKindStatusChanged     NotificationKind = "status_changed"
	NotificationKindDueAtChanged      NotificationKind = "due_at_changed"
	NotificationKindInvitationLegacy  NotificationKind = "invitation"
)
