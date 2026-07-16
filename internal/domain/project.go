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

func validateTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if utf8.RuneCountInString(title) > 100 {
		return "", ErrTitleTooLong
	}
	if utf8.RuneCountInString(title) == 0 {
		return "", ErrTitleEmpty
	}
	return title, nil
}

func validateGoal(goal string) (string, error) {
	goal = strings.TrimSpace(goal)
	if utf8.RuneCountInString(goal) > 200 {
		return "", ErrGoalTooLong
	}
	if utf8.RuneCountInString(goal) == 0 {
		return "", ErrGoalEmpty
	}
	return goal, nil
}

func NewProject(userID int, title string, goal string) (*Project, error) {

	title, err := validateTitle(title)
	if err != nil {
		return nil, err
	}

	goal, err = validateGoal(goal)
	if err != nil {
		return nil, err
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

func (p *Project) Update(title string, goal string) error {

	title, err := validateTitle(title)
	if err != nil {
		return err
	}

	goal, err = validateGoal(goal)
	if err != nil {
		return err
	}

	p.Title = title
	p.Goal = goal
	t := time.Now()
	p.UpdatedAt = &t
	return nil
}
