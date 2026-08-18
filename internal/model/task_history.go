package model

import "time"

type TaskHistory struct {
	ID        string    `json:"id" db:"id"`
	TaskID    string    `json:"task_id" db:"task_id"`
	ChangedBy string    `json:"changed_by" db:"changed_by"`
	Changes   string    `json:"changes" db:"changes"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
