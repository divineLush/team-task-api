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
