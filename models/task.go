package models

import (
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
)

type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
)

type Task struct {
	ID          uuid.UUID    `db:"id" json:"id"`
	Title       string       `db:"title" json:"title"`
	Description string       `db:"description" json:"description"`
	Status      TaskStatus   `db:"status" json:"status"`
	Priority    TaskPriority `db:"priority" json:"priority"`

	ProjectID  uuid.UUID  `db:"project_id" json:"project_id"`
	AssigneeID *uuid.UUID `db:"assignee_id" json:"assignee_id"`
	CreatedBy  uuid.UUID  `db:"created_by" json:"created_by"`
	DueDate    *time.Time `db:"due_date" json:"due_date"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

