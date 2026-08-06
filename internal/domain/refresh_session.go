package domain

import "time"

type RefreshSession struct {
	SessionID         string
	UserID            int
	TokenHash         string
	CreatedAt         time.Time
	AbsoluteExpiresAt time.Time
}
