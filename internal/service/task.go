package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/sync/singleflight"

	"github.com/team-task-api/internal/cache"
	"github.com/team-task-api/internal/model"
	"github.com/team-task-api/internal/repository"
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
	db              *sql.DB
	taskRepo        *repository.TaskRepository
	teamMemberRepo  *repository.TeamMemberRepository
	historyService  *TaskHistoryService
	taskCache       *cache.TaskCache
	listFlight      singleflight.Group
}

func NewTaskService(db *sql.DB, taskRepo *repository.TaskRepository, teamMemberRepo *repository.TeamMemberRepository, historyService *TaskHistoryService, taskCache *cache.TaskCache) *TaskService {
	return &TaskService{db: db, taskRepo: taskRepo, teamMemberRepo: teamMemberRepo, historyService: historyService, taskCache: taskCache}
}

func (s *TaskService) Create(ctx context.Context, userID string, req *model.CreateTaskRequest) (*model.Task, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	txMemberRepo := repository.NewTeamMemberRepository(tx)
	txTaskRepo := repository.NewTaskRepository(tx)

	member, err := txMemberRepo.Get(ctx, req.TeamID, userID)
	if err != nil {
		return nil, fmt.Errorf("check membership: %w", err)
	}
	if member == nil {
		return nil, ErrNotTeamMember
	}

	task := &model.Task{
		TeamID:      req.TeamID,
		Title:       req.Title,
		Description: req.Description,
		Status:      model.StatusPending,
		CreatedBy:   userID,
		AssigneeID:  req.AssigneeID,
	}

	if err := txTaskRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	if s.taskCache != nil {
		if err := s.taskCache.InvalidateTeam(ctx, req.TeamID); err != nil {
			fmt.Printf("cache invalidate error: %v\n", err)
		}
	}

	return task, nil
}

func (s *TaskService) List(ctx context.Context, userID string, filter repository.TaskFilter) ([]model.Task, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	txMemberRepo := repository.NewTeamMemberRepository(tx)

	teams, err := txMemberRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list user teams: %w", err)
	}
	if len(teams) == 0 {
		return []model.Task{}, nil
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
				return nil, ErrNotTeamMember
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

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
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
			fmt.Printf("cache get error: %v\n", err)
		}
		if cached != nil {
			return cached, nil
		}
	}

	flightKey := "list:" + cacheKey.String()
	val, err, _ := s.listFlight.Do(flightKey, func() (any, error) {
		tasks, err := s.taskRepo.List(ctx, filter)
		if err != nil {
			return nil, fmt.Errorf("list tasks: %w", err)
		}

		if s.taskCache != nil && len(tasks) > 0 {
			if err := s.taskCache.Set(ctx, cacheKey, tasks); err != nil {
				fmt.Printf("cache set error: %v\n", err)
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
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	txMemberRepo := repository.NewTeamMemberRepository(tx)
	txTaskRepo := repository.NewTaskRepository(tx)

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

	isPrivileged := member.Role == model.RoleOwner || member.Role == model.RoleAdmin
	isCreator := task.CreatedBy == userID
	isAssignee := task.AssigneeID != nil && *task.AssigneeID == userID

	if !isPrivileged && !isCreator && !isAssignee {
		return nil, ErrNotAuthorized
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
		return nil, ErrCannotReassign
	}

	if req.AssigneeID != nil {
		assignee, err := txMemberRepo.Get(ctx, task.TeamID, *req.AssigneeID)
		if err != nil {
			return nil, fmt.Errorf("check assignee membership: %w", err)
		}
		if assignee == nil {
			return nil, ErrAssigneeNotMember
		}
		task.AssigneeID = req.AssigneeID
	}

	changesJSON, err := BuildChanges(&oldTask, task)
	if err != nil {
		return nil, fmt.Errorf("build changes: %w", err)
	}

	if err := txTaskRepo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}

	if changesJSON != "" {
		txHistoryRepo := repository.NewTaskHistoryRepository(tx)
		history := &model.TaskHistory{
			TaskID:    taskID,
			ChangedBy: userID,
			Changes:   changesJSON,
		}
		if err := txHistoryRepo.Create(ctx, history); err != nil {
			return nil, fmt.Errorf("create history: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	if s.taskCache != nil {
		if err := s.taskCache.InvalidateTeam(ctx, task.TeamID); err != nil {
			fmt.Printf("cache invalidate error: %v\n", err)
		}
	}

	return task, nil
}

func (s *TaskService) GetByID(ctx context.Context, userID, taskID string) (*model.Task, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	txMemberRepo := repository.NewTeamMemberRepository(tx)
	txTaskRepo := repository.NewTaskRepository(tx)

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

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return task, nil
}

func (s *TaskService) Delete(ctx context.Context, userID, taskID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	txMemberRepo := repository.NewTeamMemberRepository(tx)
	txTaskRepo := repository.NewTaskRepository(tx)

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

	if err := txTaskRepo.Delete(ctx, taskID); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	if s.taskCache != nil {
		if err := s.taskCache.InvalidateTeam(ctx, task.TeamID); err != nil {
			fmt.Printf("cache invalidate error: %v\n", err)
		}
	}

	return nil
}
