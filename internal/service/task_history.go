package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/team-task-api/internal/model"
	"github.com/team-task-api/pkg/database"
)

type TaskHistoryService struct {
	txm                database.TxManager
	historyRepo        TaskHistoryRepository
	taskRepo           TaskRepository
	newTaskRepoInTx    func(database.Querier) TaskRepository
	newMemberRepoInTx  func(database.Querier) TeamMemberRepository
	newHistoryRepoInTx func(database.Querier) TaskHistoryRepository
}

func NewTaskHistoryService(txm database.TxManager, historyRepo TaskHistoryRepository, taskRepo TaskRepository) *TaskHistoryService {
	return &TaskHistoryService{
		txm:                txm,
		historyRepo:        historyRepo,
		taskRepo:           taskRepo,
		newTaskRepoInTx:    newTxTaskRepo,
		newMemberRepoInTx:  newTxTeamMemberRepo,
		newHistoryRepoInTx: newTxTaskHistoryRepo,
	}
}

func (s *TaskHistoryService) List(ctx context.Context, userID, taskID string) ([]model.TaskHistory, error) {
	var history []model.TaskHistory

	err := s.txm.InTx(ctx, func(tx database.Querier) error {
		txTaskRepo := s.newTaskRepoInTx(tx)
		txMemberRepo := s.newMemberRepoInTx(tx)
		txHistoryRepo := s.newHistoryRepoInTx(tx)

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

		history, err = txHistoryRepo.ListByTask(ctx, taskID)
		if err != nil {
			return fmt.Errorf("list history: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
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
