package domain

import (
	"encoding/json"
	"time"
)

type Notification struct {
	ID          int64
	RecipientID int
	ProjectID   int
	TaskID      *int
	InviteID    *int
	Kind        string
	Payload     json.RawMessage
	CreatedAt   time.Time
	ReadAt      *time.Time
}
