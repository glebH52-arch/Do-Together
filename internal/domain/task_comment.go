package domain

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
	"uuid"
)

var ErrInvalidCommentBody = errors.New("invalid comment body")
var ErrInvalidCommentPage = errors.New("invalid comment page")

type TaskComment struct {
	ID         uuid.UUID
	TaskID     uuid.UUID
	AuthorID   uuid.UUID
	AuthorName string
	Body       string
	CreatedAt  time.Time
	UpdatedAt  *time.Time
}

func ValidateCommentBody(body string) (string, error) {
	body = strings.TrimSpace(body)
	if body == "" || utf8.RuneCountInString(body) > 500 {
		return "", ErrInvalidCommentBody
	}
	return body, nil
}
