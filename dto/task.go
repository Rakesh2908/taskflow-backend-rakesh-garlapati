package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateTaskRequest struct {
	Title       string     `json:"title" validate:"required"`
	Description string     `json:"description"`
	Priority    string     `json:"priority" validate:"required,oneof=low medium high"`
	AssigneeID  *uuid.UUID `json:"assignee_id"`
	DueDate     *time.Time `json:"due_date"`
}

type UpdateTaskRequest struct {
	Title       *string    `json:"title" validate:"omitempty"`
	Description *string    `json:"description" validate:"omitempty"`
	Status      *string    `json:"status" validate:"omitempty,oneof=todo in_progress done"`
	Priority    *string    `json:"priority" validate:"omitempty,oneof=low medium high"`
	AssigneeID  *uuid.UUID `json:"assignee_id"`
	DueDate     *time.Time `json:"due_date"`
}

