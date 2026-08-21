package service

import (
	"context"

	"github.com/team-task-api/internal/model"
	"github.com/team-task-api/internal/repository"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
}

type TeamRepository interface {
	Create(ctx context.Context, team *model.Team) error
	GetByID(ctx context.Context, id string) (*model.Team, error)
	ListByUser(ctx context.Context, userID string) ([]model.Team, error)
	Update(ctx context.Context, team *model.Team) error
	Delete(ctx context.Context, id string) error
}

type TeamMemberRepository interface {
	Add(ctx context.Context, m *model.TeamMember) error
	Get(ctx context.Context, teamID, userID string) (*model.TeamMember, error)
	ListByUser(ctx context.Context, userID string) ([]model.TeamMember, error)
	GetStats(ctx context.Context, teamID string) (*model.TeamStats, error)
}

type TaskRepository interface {
	Create(ctx context.Context, task *model.Task) error
	GetByID(ctx context.Context, id string) (*model.Task, error)
	List(ctx context.Context, filter repository.TaskFilter) ([]model.Task, error)
	Update(ctx context.Context, task *model.Task) error
	Delete(ctx context.Context, id string) error
}

type TaskHistoryRepository interface {
	Create(ctx context.Context, h *model.TaskHistory) error
	ListByTask(ctx context.Context, taskID string) ([]model.TaskHistory, error)
}

type TaskCommentRepository interface {
	Create(ctx context.Context, c *model.TaskComment) error
	GetByID(ctx context.Context, taskID, commentID string) (*model.TaskComment, error)
	ListByTask(ctx context.Context, taskID string) ([]model.TaskComment, error)
	Delete(ctx context.Context, id string) error
}
