package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/team-task-api/internal/model"
)

type TeamRepository struct {
	db *sql.DB
}

func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) Create(ctx context.Context, team *model.Team) error {
	team.ID = uuid.New().String()
	query := `INSERT INTO teams (id, name, created_by) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, team.ID, team.Name, team.CreatedBy)
	if err != nil {
		return fmt.Errorf("insert team: %w", err)
	}
	return nil
}

func (r *TeamRepository) GetByID(ctx context.Context, id string) (*model.Team, error) {
	query := `SELECT id, name, created_by, created_at FROM teams WHERE id = ?`
	team := &model.Team{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&team.ID, &team.Name, &team.CreatedBy, &team.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get team by id: %w", err)
	}
	return team, nil
}

func (r *TeamRepository) List(ctx context.Context) ([]model.Team, error) {
	query := `SELECT id, name, created_by, created_at FROM teams ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list teams: %w", err)
	}
	defer rows.Close()

	var teams []model.Team
	for rows.Next() {
		var t model.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan team: %w", err)
		}
		teams = append(teams, t)
	}
	return teams, nil
}

func (r *TeamRepository) ListByUser(ctx context.Context, userID string) ([]model.Team, error) {
	query := `SELECT t.id, t.name, t.created_by, t.created_at FROM teams t
		INNER JOIN team_members tm ON tm.team_id = t.id
		WHERE tm.user_id = ?
		ORDER BY t.created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list teams by user: %w", err)
	}
	defer rows.Close()

	var teams []model.Team
	for rows.Next() {
		var t model.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan team: %w", err)
		}
		teams = append(teams, t)
	}
	return teams, nil
}

func (r *TeamRepository) Update(ctx context.Context, team *model.Team) error {
	query := `UPDATE teams SET name = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, team.Name, team.ID)
	if err != nil {
		return fmt.Errorf("update team: %w", err)
	}
	return nil
}

func (r *TeamRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM teams WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete team: %w", err)
	}
	return nil
}
