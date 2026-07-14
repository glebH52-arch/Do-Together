package handler

import (
	"time"
)

type createProjectRequest struct {
	// CreatedBy, Title, Goal и JSON-теги
	CreatedBy int    `json:"created_by"`
	Title     string `json:"title"`
	Goal      string `json:"goal"`
}

type createProjectResponse struct {
	// ID, CreatedBy, Title, Goal, Status, CreatedAt и JSON-теги
	ID        int       `json:"id"`
	CreatedBy int       `json:"created_by"`
	Title     string    `json:"title"`
	Goal      string    `json:"goal"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
