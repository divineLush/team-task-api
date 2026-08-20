package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/team-task-api/internal/model"
	"github.com/team-task-api/pkg/database"
)

type TaskCommentRepository struct {
	db database.Querier
}

func NewTaskCommentRepository(db database.Querier) *TaskCommentRepository {
	return &TaskCommentRepository{db: db}
}

func (r *TaskCommentRepository) Create(ctx context.Context, c *model.TaskComment) error {
	c.ID = uuid.New().String()
	query := `INSERT INTO task_comments (id, task_id, user_id, content) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, c.ID, c.TaskID, c.UserID, c.Content)
	if err != nil {
		return fmt.Errorf("insert task comment: %w", err)
	}
	return nil
}

func (r *TaskCommentRepository) GetByID(ctx context.Context, taskID, commentID string) (*model.TaskComment, error) {
	query := `SELECT id, task_id, user_id, content, created_at FROM task_comments
		WHERE id = ? AND task_id = ?`
	c := &model.TaskComment{}
	err := r.db.QueryRowContext(ctx, query, commentID, taskID).Scan(
		&c.ID, &c.TaskID, &c.UserID, &c.Content, &c.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get comment by id: %w", err)
	}
	return c, nil
}

func (r *TaskCommentRepository) ListByTask(ctx context.Context, taskID string) ([]model.TaskComment, error) {
	query := `SELECT id, task_id, user_id, content, created_at FROM task_comments
		WHERE task_id = ? ORDER BY created_at ASC`
	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("list task comments: %w", err)
	}
	defer rows.Close()

	var comments []model.TaskComment
	for rows.Next() {
		var c model.TaskComment
		if err := rows.Scan(&c.ID, &c.TaskID, &c.UserID, &c.Content, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan task comment: %w", err)
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func (r *TaskCommentRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM task_comments WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete task comment: %w", err)
	}
	return nil
}
