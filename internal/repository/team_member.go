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

func (r *TeamMemberRepository) GetStats(ctx context.Context, teamID string) (*model.TeamStats, error) {
	stats := &model.TeamStats{}

	// Top members by done tasks (last 30 days)
	memberQuery := `SELECT u.id, u.name, u.email, COUNT(t.id) AS done_count
		FROM team_members tm
		JOIN users u ON u.id = tm.user_id
		LEFT JOIN tasks t ON t.assignee_id = tm.user_id AND t.team_id = tm.team_id
			AND t.status = 'done' AND t.closed_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)
		WHERE tm.team_id = ?
		GROUP BY tm.user_id, u.name, u.email
		ORDER BY done_count DESC
		LIMIT 3`
	rows, err := r.db.QueryContext(ctx, memberQuery, teamID)
	if err != nil {
		return nil, fmt.Errorf("query top members: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var m model.MemberStats
		if err := rows.Scan(&m.UserID, &m.Name, &m.Email, &m.DoneCount); err != nil {
			return nil, fmt.Errorf("scan member stats: %w", err)
		}
		stats.TopMembers = append(stats.TopMembers, m)
	}
	if stats.TopMembers == nil {
		stats.TopMembers = []model.MemberStats{}
	}

	// Average close time
	avgQuery := `SELECT AVG(TIMESTAMPDIFF(SECOND, created_at, closed_at)) / 3600.0
		FROM tasks WHERE team_id = ? AND status = 'done' AND closed_at IS NOT NULL`
	var avgSecs sql.NullFloat64
	if err := r.db.QueryRowContext(ctx, avgQuery, teamID).Scan(&avgSecs); err != nil {
		return nil, fmt.Errorf("query avg close time: %w", err)
	}
	if avgSecs.Valid {
		hours := avgSecs.Float64
		stats.AvgCloseHrs = &hours
	}

	// Comment count
	commentQuery := `SELECT COUNT(*) FROM task_comments c
		JOIN tasks t ON t.id = c.task_id
		WHERE t.team_id = ?`
	if err := r.db.QueryRowContext(ctx, commentQuery, teamID).Scan(&stats.CommentCount); err != nil {
		return nil, fmt.Errorf("query comment count: %w", err)
	}

	// Tasks by status
	statusQuery := `SELECT status, COUNT(*) FROM tasks WHERE team_id = ? GROUP BY status`
	statusRows, err := r.db.QueryContext(ctx, statusQuery, teamID)
	if err != nil {
		return nil, fmt.Errorf("query task statuses: %w", err)
	}
	defer statusRows.Close()

	for statusRows.Next() {
		var ts model.TaskStatusCount
		if err := statusRows.Scan(&ts.Status, &ts.Count); err != nil {
			return nil, fmt.Errorf("scan task status: %w", err)
		}
		stats.Tasks = append(stats.Tasks, ts)
	}
	if stats.Tasks == nil {
		stats.Tasks = []model.TaskStatusCount{}
	}

	return stats, nil
}
