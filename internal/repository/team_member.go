package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/team-task-api/internal/model"
)

type TeamMemberRepository struct {
	db *sql.DB
}

func NewTeamMemberRepository(db *sql.DB) *TeamMemberRepository {
	return &TeamMemberRepository{db: db}
}

func (r *TeamMemberRepository) Add(ctx context.Context, m *model.TeamMember) error {
	query := `INSERT INTO team_members (team_id, user_id, role) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, m.TeamID, m.UserID, m.Role)
	if err != nil {
		return fmt.Errorf("add team member: %w", err)
	}
	return nil
}

func (r *TeamMemberRepository) Get(ctx context.Context, teamID, userID string) (*model.TeamMember, error) {
	query := `SELECT team_id, user_id, role FROM team_members WHERE team_id = ? AND user_id = ?`
	m := &model.TeamMember{}
	err := r.db.QueryRowContext(ctx, query, teamID, userID).Scan(
		&m.TeamID, &m.UserID, &m.Role,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get team member: %w", err)
	}
	return m, nil
}

func (r *TeamMemberRepository) ListByTeam(ctx context.Context, teamID string) ([]model.TeamMember, error) {
	query := `SELECT team_id, user_id, role FROM team_members WHERE team_id = ?`
	rows, err := r.db.QueryContext(ctx, query, teamID)
	if err != nil {
		return nil, fmt.Errorf("list team members: %w", err)
	}
	defer rows.Close()

	var members []model.TeamMember
	for rows.Next() {
		var m model.TeamMember
		if err := rows.Scan(&m.TeamID, &m.UserID, &m.Role); err != nil {
			return nil, fmt.Errorf("scan team member: %w", err)
		}
		members = append(members, m)
	}
	return members, nil
}

func (r *TeamMemberRepository) ListByUser(ctx context.Context, userID string) ([]model.TeamMember, error) {
	query := `SELECT team_id, user_id, role FROM team_members WHERE user_id = ?`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list user teams: %w", err)
	}
	defer rows.Close()

	var members []model.TeamMember
	for rows.Next() {
		var m model.TeamMember
		if err := rows.Scan(&m.TeamID, &m.UserID, &m.Role); err != nil {
			return nil, fmt.Errorf("scan team member: %w", err)
		}
		members = append(members, m)
	}
	return members, nil
}

func (r *TeamMemberRepository) UpdateRole(ctx context.Context, teamID, userID string, role model.TeamRole) error {
	query := `UPDATE team_members SET role = ? WHERE team_id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, role, teamID, userID)
	if err != nil {
		return fmt.Errorf("update team member role: %w", err)
	}
	return nil
}

func (r *TeamMemberRepository) Remove(ctx context.Context, teamID, userID string) error {
	query := `DELETE FROM team_members WHERE team_id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, teamID, userID)
	if err != nil {
		return fmt.Errorf("remove team member: %w", err)
	}
	return nil
}
