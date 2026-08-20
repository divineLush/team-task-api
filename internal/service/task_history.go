package service

import (
	"context"
	"database/sql"
	"encoding/json"
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

func (s *TaskHistoryService) Create(ctx context.Context, history *model.TaskHistory) error {
	return s.historyRepo.Create(ctx, history)
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

type FieldChange struct {
	Old any `json:"old"`
	New any `json:"new"`
}

func BuildChanges(old, new *model.Task) (string, error) {
	changes := make(map[string]FieldChange)

	if old.Title != new.Title {
		changes["title"] = FieldChange{Old: old.Title, New: new.Title}
	}
	if old.Description != new.Description {
		changes["description"] = FieldChange{Old: old.Description, New: new.Description}
	}
	if old.Status != new.Status {
		changes["status"] = FieldChange{Old: old.Status, New: new.Status}
	}
	if !ptrEqual(old.AssigneeID, new.AssigneeID) {
		changes["assignee_id"] = FieldChange{Old: old.AssigneeID, New: new.AssigneeID}
	}

	if len(changes) == 0 {
		return "", nil
	}

	b, err := json.Marshal(changes)
	if err != nil {
		return "", fmt.Errorf("marshal changes: %w", err)
	}
	return string(b), nil
}

func ptrEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
