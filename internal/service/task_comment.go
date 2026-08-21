package service

import (
	"context"
	"fmt"

	"github.com/team-task-api/internal/model"
	"github.com/team-task-api/pkg/database"
)

type TaskCommentService struct {
	txm                database.TxManager
	newTaskRepoInTx    func(database.Querier) TaskRepository
	newMemberRepoInTx  func(database.Querier) TeamMemberRepository
	newCommentRepoInTx func(database.Querier) TaskCommentRepository
}

func NewTaskCommentService(txm database.TxManager, commentRepo TaskCommentRepository, taskRepo TaskRepository) *TaskCommentService {
	return &TaskCommentService{
		txm:                txm,
		newTaskRepoInTx:    newTxTaskRepo,
		newMemberRepoInTx:  newTxTeamMemberRepo,
		newCommentRepoInTx: newTxTaskCommentRepo,
	}
}

func (s *TaskCommentService) Create(ctx context.Context, userID, taskID string, req *model.CreateCommentRequest) (*model.TaskComment, error) {
	var comment *model.TaskComment

	err := s.txm.InTx(ctx, func(tx database.Querier) error {
		txTaskRepo := s.newTaskRepoInTx(tx)
		txMemberRepo := s.newMemberRepoInTx(tx)
		txCommentRepo := s.newCommentRepoInTx(tx)

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

		comment = &model.TaskComment{
			TaskID:  taskID,
			UserID:  userID,
			Content: req.Content,
		}

		if err := txCommentRepo.Create(ctx, comment); err != nil {
			return fmt.Errorf("create comment: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *TaskCommentService) List(ctx context.Context, userID, taskID string) ([]model.TaskComment, error) {
	var comments []model.TaskComment

	err := s.txm.InTx(ctx, func(tx database.Querier) error {
		txTaskRepo := s.newTaskRepoInTx(tx)
		txMemberRepo := s.newMemberRepoInTx(tx)
		txCommentRepo := s.newCommentRepoInTx(tx)

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

		comments, err = txCommentRepo.ListByTask(ctx, taskID)
		if err != nil {
			return fmt.Errorf("list comments: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return comments, nil
}

func (s *TaskCommentService) Delete(ctx context.Context, userID, taskID, commentID string) error {
	return s.txm.InTx(ctx, func(tx database.Querier) error {
		txTaskRepo := s.newTaskRepoInTx(tx)
		txMemberRepo := s.newMemberRepoInTx(tx)
		txCommentRepo := s.newCommentRepoInTx(tx)

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

		existing, err := txCommentRepo.GetByID(ctx, taskID, commentID)
		if err != nil {
			return fmt.Errorf("get comment: %w", err)
		}
		if existing == nil {
			return ErrCommentNotFound
		}

		if err := txCommentRepo.Delete(ctx, commentID); err != nil {
			return fmt.Errorf("delete comment: %w", err)
		}

		return nil
	})
}
