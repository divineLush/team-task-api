package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/team-task-api/internal/model"
	"github.com/team-task-api/pkg/database"
)

type TeamMemberRepository struct {
	db database.Querier
}

func NewTeamMemberRepository(db database.Querier) *TeamMemberRepository {
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
