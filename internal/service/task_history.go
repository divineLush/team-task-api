package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/team-task-api/internal/model"
	"github.com/team-task-api/internal/repository"
)

type TaskHistoryService struct {
	db             *sql.DB
	historyRepo    *repository.TaskHistoryRepository
	taskRepo       *repository.TaskRepository
	teamMemberRepo *repository.TeamMemberRepository
}

func NewTaskHistoryService(db *sql.DB, historyRepo *repository.TaskHistoryRepository, taskRepo *repository.TaskRepository, teamMemberRepo *repository.TeamMemberRepository) *TaskHistoryService {
	return &TaskHistoryService{db: db, historyRepo: historyRepo, taskRepo: taskRepo, teamMemberRepo: teamMemberRepo}
}

func (s *TaskHistoryService) List(ctx context.Context, userID, taskID string) ([]model.TaskHistory, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	txTaskRepo := repository.NewTaskRepository(tx)
	txMemberRepo := repository.NewTeamMemberRepository(tx)
	txHistoryRepo := repository.NewTaskHistoryRepository(tx)

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

	history, err := txHistoryRepo.ListByTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("list history: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return history, nil
}
