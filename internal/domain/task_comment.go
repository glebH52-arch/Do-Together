package domain

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrInvalidCommentBody = errors.New("invalid comment body")
var ErrInvalidCommentPage = errors.New("invalid comment page")

type TaskComment struct {
	ID         int64
	TaskID     int
	AuthorID   int
	AuthorName string
	Body       string
	CreatedAt  time.Time
	UpdatedAt  *time.Time
}

func ValidateCommentBody(body string) (string, error) {
	body = strings.TrimSpace(body)
	if body == "" || utf8.RuneCountInString(body) > 5000 {
		return "", ErrInvalidCommentBody
	}
	return body, nil
}
