package model

import "time"

type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
)

type Task struct {
	ID          uint64     `json:"id" db:"id"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	Status      TaskStatus `json:"status" db:"status"`
	TeamID      uint64     `json:"team_id" db:"team_id"`
	AssignedTo  *uint64    `json:"assigned_to,omitempty" db:"assigned_to"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateTaskRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	TeamID      uint64     `json:"team_id"`
	AssignedTo  *uint64    `json:"assigned_to,omitempty"`
	Status      TaskStatus `json:"status,omitempty"`
}

type UpdateTaskRequest struct {
	Title       *string     `json:"title,omitempty"`
	Description *string     `json:"description,omitempty"`
	Status      *TaskStatus `json:"status,omitempty"`
	AssignedTo  *uint64     `json:"assigned_to,omitempty"`
}
