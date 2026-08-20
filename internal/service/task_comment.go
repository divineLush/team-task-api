package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/team-task-api/internal/model"
	"github.com/team-task-api/internal/repository"
)

type TaskCommentService struct {
	db             *sql.DB
	commentRepo    *repository.TaskCommentRepository
	taskRepo       *repository.TaskRepository
	teamMemberRepo *repository.TeamMemberRepository
}

func NewTaskCommentService(db *sql.DB, commentRepo *repository.TaskCommentRepository, taskRepo *repository.TaskRepository, teamMemberRepo *repository.TeamMemberRepository) *TaskCommentService {
	return &TaskCommentService{db: db, commentRepo: commentRepo, taskRepo: taskRepo, teamMemberRepo: teamMemberRepo}
}

func (s *TaskCommentService) Create(ctx context.Context, userID, taskID string, req *model.CreateCommentRequest) (*model.TaskComment, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	txTaskRepo := repository.NewTaskRepository(tx)
	txMemberRepo := repository.NewTeamMemberRepository(tx)
	txCommentRepo := repository.NewTaskCommentRepository(tx)

	task, err := txTaskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	member, err := txMemberRepo.Get(ctx, task.TeamID, userID)
	if err != nil {
		return nil, fmt.Errorf("check membership: %w", err)
	}
	if member == nil {
		return nil, ErrNotTeamMember
	}

	comment := &model.TaskComment{
		TaskID:  taskID,
		UserID:  userID,
		Content: req.Content,
	}

	if err := txCommentRepo.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return comment, nil
}

func (s *TaskCommentService) List(ctx context.Context, userID, taskID string) ([]model.TaskComment, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	txTaskRepo := repository.NewTaskRepository(tx)
	txMemberRepo := repository.NewTeamMemberRepository(tx)
	txCommentRepo := repository.NewTaskCommentRepository(tx)

	task, err := txTaskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	member, err := txMemberRepo.Get(ctx, task.TeamID, userID)
	if err != nil {
		return nil, fmt.Errorf("check membership: %w", err)
	}
	if member == nil {
		return nil, ErrNotTeamMember
	}

	comments, err := txCommentRepo.ListByTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return comments, nil
}

func (s *TaskCommentService) Delete(ctx context.Context, userID, taskID, commentID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	txTaskRepo := repository.NewTaskRepository(tx)
	txMemberRepo := repository.NewTeamMemberRepository(tx)
	txCommentRepo := repository.NewTaskCommentRepository(tx)

	task, err := txTaskRepo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("get task: %w", err)
	}
	if task == nil {
		return ErrTaskNotFound
	}

	member, err := txMemberRepo.Get(ctx, task.TeamID, userID)
	if err != nil {
		return fmt.Errorf("check membership: %w", err)
	}
	if member == nil {
		return ErrNotTeamMember
	}

	if err := txCommentRepo.Delete(ctx, commentID); err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
