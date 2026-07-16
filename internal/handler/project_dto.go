package handler

import (
	"do-together/internal/domain"
	"time"
)

type updateProjectRequest struct {
	Title string `json:"title"`
	Goal  string `json:"goal"`
}

type projectRequest struct {
	CreatedBy int    `json:"created_by"`
	Title     string `json:"title"`
	Goal      string `json:"goal"`
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
