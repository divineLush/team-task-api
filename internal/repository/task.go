package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/team-task-api/internal/model"
	"github.com/team-task-api/pkg/database"
)

var ErrConcurrentUpdate = errors.New("concurrent update: task was modified by another user")

type TaskRepository struct {
	db database.Querier
}

func NewTaskRepository(db database.Querier) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *model.Task) error {
	task.ID = uuid.New().String()
	query := `INSERT INTO tasks (id, team_id, title, description, status, created_by, assignee_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		task.ID, task.TeamID, task.Title, task.Description, task.Status, task.CreatedBy, task.AssigneeID,
	)
	if err != nil {
		return fmt.Errorf("insert task: %w", err)
	}
	return nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id string) (*model.Task, error) {
	query := `SELECT id, team_id, title, description, status, created_by, assignee_id,
		created_at, updated_at, closed_at, version FROM tasks WHERE id = ?`
	t := &model.Task{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.TeamID, &t.Title, &t.Description, &t.Status, &t.CreatedBy, &t.AssigneeID,
		&t.CreatedAt, &t.UpdatedAt, &t.ClosedAt, &t.Version,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get task by id: %w", err)
	}
	return t, nil
}

func (r *TaskRepository) ListByTeam(ctx context.Context, teamID string) ([]model.Task, error) {
	query := `SELECT id, team_id, title, description, status, created_by, assignee_id,
		created_at, updated_at, closed_at, version FROM tasks WHERE team_id = ? ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, teamID)
	if err != nil {
		return nil, fmt.Errorf("list tasks by team: %w", err)
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(
			&t.ID, &t.TeamID, &t.Title, &t.Description, &t.Status, &t.CreatedBy, &t.AssigneeID,
			&t.CreatedAt, &t.UpdatedAt, &t.ClosedAt, &t.Version,
		); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *TaskRepository) ListByAssignee(ctx context.Context, assigneeID string) ([]model.Task, error) {
	query := `SELECT id, team_id, title, description, status, created_by, assignee_id,
		created_at, updated_at, closed_at, version FROM tasks WHERE assignee_id = ? ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, assigneeID)
	if err != nil {
		return nil, fmt.Errorf("list tasks by assignee: %w", err)
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(
			&t.ID, &t.TeamID, &t.Title, &t.Description, &t.Status, &t.CreatedBy, &t.AssigneeID,
			&t.CreatedAt, &t.UpdatedAt, &t.ClosedAt, &t.Version,
		); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

type TaskFilter struct {
	TeamIDs   []string
	Status    *string
	AssigneeID *string
	Limit     int
	Offset    int
}

func (r *TaskRepository) List(ctx context.Context, filter TaskFilter) ([]model.Task, error) {
	placeholders := make([]string, len(filter.TeamIDs))
	args := make([]any, len(filter.TeamIDs))
	for i, id := range filter.TeamIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	query := `SELECT id, team_id, title, description, status, created_by, assignee_id,
		created_at, updated_at, closed_at, version FROM tasks WHERE team_id IN (` + strings.Join(placeholders, ",") + `)`

	if filter.Status != nil {
		query += ` AND status = ?`
		args = append(args, *filter.Status)
	}
	if filter.AssigneeID != nil {
		query += ` AND assignee_id = ?`
		args = append(args, *filter.AssigneeID)
	}

	query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(
			&t.ID, &t.TeamID, &t.Title, &t.Description, &t.Status, &t.CreatedBy, &t.AssigneeID,
			&t.CreatedAt, &t.UpdatedAt, &t.ClosedAt, &t.Version,
		); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *model.Task) error {
	query := `UPDATE tasks SET title = ?, description = ?, status = ?, assignee_id = ?, version = version + 1
		WHERE id = ? AND version = ?`
	result, err := r.db.ExecContext(ctx, query,
		task.Title, task.Description, task.Status, task.AssigneeID, task.ID, task.Version,
	)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return ErrConcurrentUpdate
	}
	return nil
}

func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tasks WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	return nil
}
