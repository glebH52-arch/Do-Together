package domain

import (
	"time"
	"uuid"
)

type RefreshSession struct {
	SessionID         string
	UserID            uuid.UUID
	TokenHash         string
	CreatedAt         time.Time
	AbsoluteExpiresAt time.Time
}
