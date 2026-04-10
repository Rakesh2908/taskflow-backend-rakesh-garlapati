package dto

import "github.com/google/uuid"

type CreateProjectRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type UpdateProjectRequest struct {
	Name        *string `json:"name" validate:"omitempty"`
	Description *string `json:"description" validate:"omitempty"`
}

type ProjectStatsResponse struct {
	ProjectID  uuid.UUID                       `json:"project_id"`
	ByStatus   map[string]int                  `json:"by_status"`
	ByAssignee map[string]map[string]int       `json:"by_assignee"`
}

