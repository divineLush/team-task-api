package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/team-task-api/internal/model"
	"github.com/team-task-api/internal/repository"
)

var (
	ErrNotTeamMember = errors.New("user is not a member of this team")
)

type TaskService struct {
	db             *sql.DB
	taskRepo       *repository.TaskRepository
	teamMemberRepo *repository.TeamMemberRepository
}

func NewTaskService(db *sql.DB, taskRepo *repository.TaskRepository, teamMemberRepo *repository.TeamMemberRepository) *TaskService {
	return &TaskService{db: db, taskRepo: taskRepo, teamMemberRepo: teamMemberRepo}
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

	return task, nil
}

func (s *TaskService) List(ctx context.Context, userID string, filter repository.TaskFilter) ([]model.Task, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	txMemberRepo := repository.NewTeamMemberRepository(tx)
	txTaskRepo := repository.NewTaskRepository(tx)

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

	tasks, err := txTaskRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return tasks, nil
}
