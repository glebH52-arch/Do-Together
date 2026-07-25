package handler

import (
	"do-together/internal/domain"
	"do-together/internal/repository"
	"errors"
	"net/url"
	"strconv"
	"time"
)

type updateProjectRequest struct {
	Title *string `json:"title"`
	Goal  *string `json:"goal"`
}

type projectRequest struct {
	Title string `json:"title"`
	Goal  string `json:"goal"`
}

type projectResponse struct {
	ID        int        `json:"id"`
	CreatedBy int        `json:"created_by"`
	Title     string     `json:"title"`
	Goal      string     `json:"goal"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

func projectToResponse(project *domain.Project) projectResponse {
	return projectResponse{
		ID:        project.ID,
		CreatedBy: project.CreatedBy,
		Title:     project.Title,
		Goal:      project.Goal,
		Status:    string(project.Status),
		CreatedAt: project.CreatedAt,
		UpdatedAt: project.UpdatedAt,
	}
}

func parseProjectListOptions(query url.Values) (repository.ProjectListOptions, error) {
	option := repository.ProjectListOptions{
		Limit:  20,
		Offset: 0,
	}
	if limit := query.Get("limit"); limit != "" {
		value, err := strconv.Atoi(limit)
		if err != nil {
			return option, errors.New("invalid limit")
		}
		if value <= 0 || value > 100 {
			return option, errors.New("invalid limit")
		}
		option.Limit = value
	}

	if offset := query.Get("offset"); offset != "" {
		value, err := strconv.Atoi(offset)
		if err != nil {
			return option, errors.New("invalid offset")
		}
		if value < 0 {
			return option, errors.New("invalid offset")
		}
		option.Offset = value
	}
	if status := query.Get("status"); status != "" {
		switch status {
		case string(domain.ProjectStatusActive):
			s := domain.ProjectStatusActive
			option.Status = &s

		case string(domain.ProjectStatusArchived):
			s := domain.ProjectStatusArchived
			option.Status = &s

		default:
			return option, errors.New("invalid project status")
		}
	}

	return option, nil
}
