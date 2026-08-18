package model

import "time"

type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
)

type Task struct {
	ID          string     `json:"id" db:"id"`
	TeamID      string     `json:"team_id" db:"team_id"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	Status      TaskStatus `json:"status" db:"status"`
	CreatedBy   string     `json:"created_by" db:"created_by"`
	AssigneeID  *string    `json:"assignee_id,omitempty" db:"assignee_id"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	ClosedAt    *time.Time `json:"closed_at,omitempty" db:"closed_at"`
	Version     int        `json:"version" db:"version"`
}

type CreateTaskRequest struct {
	TeamID      string  `json:"team_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	AssigneeID  *string `json:"assignee_id,omitempty"`
}

type UpdateTaskRequest struct {
	Title       *string     `json:"title,omitempty"`
	Description *string     `json:"description,omitempty"`
	Status      *TaskStatus `json:"status,omitempty"`
	AssigneeID  *string     `json:"assignee_id,omitempty"`
}
