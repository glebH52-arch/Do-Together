package service

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
	"uuid"
)

const (
	DefaultActivityLimit = 50
	MaxActivityLimit     = 100
)

var (
	ErrInvalidActivityCursor = errors.New("invalid activity cursor")
	ErrInvalidActivityLimit  = errors.New("activity limit must be between 1 and 100")
)

type ProjectActivityPage struct {
	Items      []*domain.ProjectActivity
	NextCursor *string
}

type ProjectActivityService struct {
	repository repository.ProjectActivityRepository
}

func NewProjectActivityService(repository repository.ProjectActivityRepository) *ProjectActivityService {
	return &ProjectActivityService{repository: repository}
}

type activityCursorPayload struct {
	OccurredAt string    `json:"occurred_at"`
	ID         uuid.UUID `json:"id"`
}

func EncodeActivityCursor(cursor repository.ActivityCursor) (string, error) {
	payload, err := json.Marshal(activityCursorPayload{
		OccurredAt: cursor.OccurredAt.UTC().Format(time.RFC3339Nano),
		ID:         cursor.ID,
	})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func DecodeActivityCursor(value string) (*repository.ActivityCursor, error) {
	if value == "" {
		return nil, nil
	}
	payload, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, ErrInvalidActivityCursor
	}
	var decoded activityCursorPayload
	if err := json.Unmarshal(payload, &decoded); err != nil || decoded.ID == uuid.Nil() {
		return nil, ErrInvalidActivityCursor
	}
	occurredAt, err := time.Parse(time.RFC3339Nano, decoded.OccurredAt)
	if err != nil {
		return nil, ErrInvalidActivityCursor
	}
	return &repository.ActivityCursor{OccurredAt: occurredAt, ID: decoded.ID}, nil
}

func (s *ProjectActivityService) List(ctx context.Context, actorID, projectID uuid.UUID, cursorText string, limit int) (*ProjectActivityPage, error) {
	if actorID == uuid.Nil() || projectID == uuid.Nil() {
		return nil, repository.ErrProjectNotFound
	}
	if limit < 1 || limit > MaxActivityLimit {
		return nil, ErrInvalidActivityLimit
	}
	cursor, err := DecodeActivityCursor(cursorText)
	if err != nil {
		return nil, err
	}
	items, err := s.repository.List(ctx, actorID, projectID, cursor, limit+1)
	if err != nil {
		return nil, err
	}
	page := &ProjectActivityPage{Items: items}
	if len(items) > limit {
		page.Items = items[:limit]
		last := page.Items[len(page.Items)-1]
		next, err := EncodeActivityCursor(repository.ActivityCursor{OccurredAt: last.OccurredAt, ID: last.ID})
		if err != nil {
			return nil, err
		}
		page.NextCursor = &next
	}
	return page, nil
}
