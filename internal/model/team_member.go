package model

type TeamRole string

const (
	RoleOwner  TeamRole = "owner"
	RoleAdmin  TeamRole = "admin"
	RoleMember TeamRole = "member"
)

type TeamMember struct {
	TeamID string   `json:"team_id" db:"team_id"`
	UserID string   `json:"user_id" db:"user_id"`
	Role   TeamRole `json:"role" db:"role"`
}

type AddTeamMemberRequest struct {
	UserID string   `json:"user_id"`
	Role   TeamRole `json:"role"`
}

type UpdateTeamMemberRequest struct {
	Role *TeamRole `json:"role,omitempty"`
}
