package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/team-task-api/internal/model"
)

type TaskHistoryRepository struct {
	db *sql.DB
}

func NewTaskHistoryRepository(db *sql.DB) *TaskHistoryRepository {
	return &TaskHistoryRepository{db: db}
}

func (r *TaskHistoryRepository) Create(ctx context.Context, h *model.TaskHistory) error {
	h.ID = uuid.New().String()
	query := `INSERT INTO task_history (id, task_id, changed_by, changes) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, h.ID, h.TaskID, h.ChangedBy, h.Changes)
	if err != nil {
		return fmt.Errorf("insert task history: %w", err)
	}
	return nil
}

func (r *TaskHistoryRepository) ListByTask(ctx context.Context, taskID string) ([]model.TaskHistory, error) {
	query := `SELECT id, task_id, changed_by, changes, created_at FROM task_history
		WHERE task_id = ? ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("list task history: %w", err)
	}
	defer rows.Close()

	var history []model.TaskHistory
	for rows.Next() {
		var h model.TaskHistory
		if err := rows.Scan(&h.ID, &h.TaskID, &h.ChangedBy, &h.Changes, &h.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan task history: %w", err)
		}
		history = append(history, h)
	}
	return history, nil
}
