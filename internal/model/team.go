package model

import "time"

type Team struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedBy string    `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateTeamRequest struct {
	Name string `json:"name"`
}

type UpdateTeamRequest struct {
	Name *string `json:"name,omitempty"`
}

type TeamStats struct {
	TopMembers  []MemberStats    `json:"top_members"`
	AvgCloseHrs *float64         `json:"avg_close_time_hours,omitempty"`
	CommentCount int             `json:"comment_count"`
	Tasks        []TaskStatusCount `json:"tasks"`
}

type MemberStats struct {
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	DoneCount   int    `json:"done_count"`
}

type TaskStatusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}
