package model

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	TeamID   uint64 `json:"team_id"`
}

type AuthResponse struct {
	Token string `json:"token"`
}
