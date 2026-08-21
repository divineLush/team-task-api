package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"golang.org/x/sync/singleflight"

	"github.com/team-task-api/internal/cache"
	"github.com/team-task-api/internal/model"
	"github.com/team-task-api/internal/repository"
	"github.com/team-task-api/pkg/database"
)

var (
	ErrNotTeamMember     = errors.New("user is not a member of this team")
	ErrTaskNotFound      = errors.New("task not found")
	ErrCommentNotFound   = errors.New("comment not found")
	ErrAssigneeNotMember = errors.New("assignee is not a member of this team")
	ErrNotAuthorized     = errors.New("only task creator or assignee can edit the task")
	ErrCannotReassign    = errors.New("assignee cannot reassign the task")
)

type TaskService struct {
	txm                database.TxManager
	historyService     *TaskHistoryService
	taskCache          *cache.TaskCache
	log                *slog.Logger
	listFlight         singleflight.Group
	newTaskRepoInTx    func(database.Querier) TaskRepository
	newMemberRepoInTx  func(database.Querier) TeamMemberRepository
	newHistoryRepoInTx func(database.Querier) TaskHistoryRepository
}

func NewTaskService(txm database.TxManager, taskRepo TaskRepository, teamMemberRepo TeamMemberRepository, historyService *TaskHistoryService, taskCache *cache.TaskCache, log *slog.Logger) *TaskService {
	if log == nil {
		log = slog.Default()
	}
	return &TaskService{
		txm:                txm,
		historyService:     historyService,
		taskCache:          taskCache,
		log:                log,
		newTaskRepoInTx:    newTxTaskRepo,
		newMemberRepoInTx:  newTxTeamMemberRepo,
		newHistoryRepoInTx: newTxTaskHistoryRepo,
	}
}

func (s *TaskService) logCacheError(msg string, err error) {
	s.log.Warn(msg, "error", err)
}

func (s *TaskService) Create(ctx context.Context, userID string, req *model.CreateTaskRequest) (*model.Task, error) {
	var task *model.Task

	err := s.txm.InTx(ctx, func(tx database.Querier) error {
		txMemberRepo := s.newMemberRepoInTx(tx)
		txTaskRepo := s.newTaskRepoInTx(tx)

		member, err := txMemberRepo.Get(ctx, req.TeamID, userID)
		if err != nil {
			return fmt.Errorf("check membership: %w", err)
		}
		if member == nil {
			return ErrNotTeamMember
		}

		task = &model.Task{
			TeamID:      req.TeamID,
			Title:       req.Title,
			Description: req.Description,
			Status:      model.StatusPending,
			CreatedBy:   userID,
			AssigneeID:  req.AssigneeID,
		}

		if err := txTaskRepo.Create(ctx, task); err != nil {
			return fmt.Errorf("create task: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if s.taskCache != nil {
		if err := s.taskCache.InvalidateTeam(ctx, req.TeamID); err != nil {
			s.logCacheError("cache invalidate error", err)
		}
	}

	return task, nil
}

func (s *TaskService) List(ctx context.Context, userID string, filter repository.TaskFilter) ([]model.Task, error) {
	err := s.txm.InTx(ctx, func(tx database.Querier) error {
		txMemberRepo := s.newMemberRepoInTx(tx)

		teams, err := txMemberRepo.ListByUser(ctx, userID)
		if err != nil {
			return fmt.Errorf("list user teams: %w", err)
		}
		if len(teams) == 0 {
			return nil
		}

		teamIDs := make([]string, len(teams))
		for i, t := range teams {
			teamIDs[i] = t.TeamID
		}

		if filter.TeamIDs != nil {
			allowed := make(map[string]bool, len(teamIDs))
			for _, id := range teamIDs {
				allowed[id] = true
			}
			for _, id := range filter.TeamIDs {
				if !allowed[id] {
					return ErrNotTeamMember
				}
			}
		} else {
			filter.TeamIDs = teamIDs
		}

		if filter.Limit <= 0 {
			filter.Limit = 50
		}
		if filter.Limit > 100 {
			filter.Limit = 100
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(filter.TeamIDs) == 0 {
		return []model.Task{}, nil
	}

	cacheKey := cache.ListKey{
		TeamIDs: filter.TeamIDs,
		Limit:   filter.Limit,
		Offset:  filter.Offset,
	}
	if filter.Status != nil {
		cacheKey.Status = *filter.Status
	}
	if filter.AssigneeID != nil {
		cacheKey.AssigneeID = *filter.AssigneeID
	}

	if s.taskCache != nil {
		cached, err := s.taskCache.Get(ctx, cacheKey)
		if err != nil {
			s.logCacheError("cache get error", err)
		}
		if cached != nil {
			return cached, nil
		}
	}

	flightKey := "list:" + cacheKey.String()
	val, err, _ := s.listFlight.Do(flightKey, func() (any, error) {
		var tasks []model.Task
		err := s.txm.InTx(ctx, func(tx database.Querier) error {
			txTaskRepo := s.newTaskRepoInTx(tx)
			var err error
			tasks, err = txTaskRepo.List(ctx, filter)
			if err != nil {
				return fmt.Errorf("list tasks: %w", err)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}

		if s.taskCache != nil && len(tasks) > 0 {
			if err := s.taskCache.Set(ctx, cacheKey, tasks); err != nil {
				s.logCacheError("cache set error", err)
			}
		}

		return tasks, nil
	})
	if err != nil {
		return nil, err
	}

	return val.([]model.Task), nil
}

func (s *TaskService) Update(ctx context.Context, userID, taskID string, req *model.UpdateTaskRequest) (*model.Task, error) {
	var task *model.Task

	err := s.txm.InTx(ctx, func(tx database.Querier) error {
		txMemberRepo := s.newMemberRepoInTx(tx)
		txTaskRepo := s.newTaskRepoInTx(tx)

		var err error
		task, err = txTaskRepo.GetByID(ctx, taskID)
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

		isPrivileged := member.Role == model.RoleOwner || member.Role == model.RoleAdmin
		isCreator := task.CreatedBy == userID
		isAssignee := task.AssigneeID != nil && *task.AssigneeID == userID

		if !isPrivileged && !isCreator && !isAssignee {
			return ErrNotAuthorized
		}

		oldTask := *task

		if req.Title != nil {
			task.Title = *req.Title
		}
		if req.Description != nil {
			task.Description = *req.Description
		}
		if req.Status != nil {
			task.Status = *req.Status
		}
		if req.AssigneeID != nil && !isPrivileged && !isCreator {
			return ErrCannotReassign
		}

		if req.AssigneeID != nil {
			assignee, err := txMemberRepo.Get(ctx, task.TeamID, *req.AssigneeID)
			if err != nil {
				return fmt.Errorf("check assignee membership: %w", err)
			}
			if assignee == nil {
				return ErrAssigneeNotMember
			}
			task.AssigneeID = req.AssigneeID
		}

		changesJSON, err := BuildChanges(&oldTask, task)
		if err != nil {
			return fmt.Errorf("build changes: %w", err)
		}

		if err := txTaskRepo.Update(ctx, task); err != nil {
			return fmt.Errorf("update task: %w", err)
		}

		if changesJSON != "" {
			txHistoryRepo := s.newHistoryRepoInTx(tx)
			history := &model.TaskHistory{
				TaskID:    taskID,
				ChangedBy: userID,
				Changes:   changesJSON,
			}
			if err := txHistoryRepo.Create(ctx, history); err != nil {
				return fmt.Errorf("create history: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if s.taskCache != nil {
		if err := s.taskCache.InvalidateTeam(ctx, task.TeamID); err != nil {
			s.logCacheError("cache invalidate error", err)
		}
	}

	return task, nil
}

func (s *TaskService) GetByID(ctx context.Context, userID, taskID string) (*model.Task, error) {
	var task *model.Task

	err := s.txm.InTx(ctx, func(tx database.Querier) error {
		txMemberRepo := s.newMemberRepoInTx(tx)
		txTaskRepo := s.newTaskRepoInTx(tx)

		var err error
		task, err = txTaskRepo.GetByID(ctx, taskID)
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

		return nil
	})
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *TaskService) Delete(ctx context.Context, userID, taskID string) error {
	var teamID string

	err := s.txm.InTx(ctx, func(tx database.Querier) error {
		txMemberRepo := s.newMemberRepoInTx(tx)
		txTaskRepo := s.newTaskRepoInTx(tx)

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

		teamID = task.TeamID

		if err := txTaskRepo.Delete(ctx, taskID); err != nil {
			return fmt.Errorf("delete task: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if s.taskCache != nil {
		if err := s.taskCache.InvalidateTeam(ctx, teamID); err != nil {
			s.logCacheError("cache invalidate error", err)
		}
	}

	return nil
}
