package domain

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrGoalTooLong  = errors.New("project goal is too long")
	ErrTitleTooLong = errors.New("project title is too long")
	ErrTitleEmpty   = errors.New("project title is empty")
	ErrGoalEmpty    = errors.New("project goal is empty")
)

type ProjectStatus string

const (
	ProjectStatusActive ProjectStatus = "active"
)

type Project struct {
	ID        int
	CreatedBy int
	Title     string
	Goal      string
	Status    ProjectStatus
	CreatedAt time.Time
	UpdatedAt *time.Time
}

func NewProject(userID int, title string, goal string) (*Project, error) {
	title = strings.TrimSpace(title)
	if utf8.RuneCountInString(title) > 100 {
		return nil, ErrTitleTooLong
	}
	if utf8.RuneCountInString(title) == 0 {
		return nil, ErrTitleEmpty
	}
	goal = strings.TrimSpace(goal)
	if utf8.RuneCountInString(goal) > 200 {
		return nil, ErrGoalTooLong
	}
	if utf8.RuneCountInString(goal) == 0 {
		return nil, ErrGoalEmpty
	}

	project := Project{
		CreatedBy: userID,
		Title:     title,
		Goal:      goal,
		Status:    ProjectStatusActive,
		CreatedAt: time.Now(),
	}
	return &project, nil
}
