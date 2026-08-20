package service

import (
	"context"
	"sync"

	"github.com/team-task-api/internal/model"
	"github.com/team-task-api/internal/repository"
	"github.com/team-task-api/pkg/database"
)

type mockTxManager struct{}

func (m *mockTxManager) InTx(ctx context.Context, fn func(tx database.Querier) error) error {
	return fn(nil)
}

type mockUserRepo struct {
	mu    sync.Mutex
	users map[string]*model.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*model.User)}
}

func (r *mockUserRepo) Create(_ context.Context, user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.users {
		if u.Email == user.Email {
			return &mockDupKeyError{}
		}
	}
	r.users[user.ID] = user
	return nil
}

func (r *mockUserRepo) GetByEmail(_ context.Context, email string) (*model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}

type mockTeamRepo struct {
	mu    sync.Mutex
	teams map[string]*model.Team
}

func newMockTeamRepo() *mockTeamRepo {
	return &mockTeamRepo{teams: make(map[string]*model.Team)}
}

func (r *mockTeamRepo) Create(_ context.Context, team *model.Team) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.teams[team.ID] = team
	return nil
}

func (r *mockTeamRepo) GetByID(_ context.Context, id string) (*model.Team, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.teams[id]
	if !ok {
		return nil, nil
	}
	return t, nil
}

func (r *mockTeamRepo) ListByUser(_ context.Context, userID string) ([]model.Team, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var teams []model.Team
	for _, t := range r.teams {
		if t.CreatedBy == userID {
			teams = append(teams, *t)
		}
	}
	return teams, nil
}

func (r *mockTeamRepo) Update(_ context.Context, team *model.Team) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.teams[team.ID] = team
	return nil
}

func (r *mockTeamRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.teams, id)
	return nil
}

type mockTeamMemberRepo struct {
	mu      sync.Mutex
	members map[string]*model.TeamMember
}

func newMockTeamMemberRepo() *mockTeamMemberRepo {
	return &mockTeamMemberRepo{members: make(map[string]*model.TeamMember)}
}

func (r *mockTeamMemberRepo) key(teamID, userID string) string {
	return teamID + ":" + userID
}

func (r *mockTeamMemberRepo) Add(_ context.Context, m *model.TeamMember) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.members[r.key(m.TeamID, m.UserID)] = m
	return nil
}

func (r *mockTeamMemberRepo) Get(_ context.Context, teamID, userID string) (*model.TeamMember, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.members[r.key(teamID, userID)]
	if !ok {
		return nil, nil
	}
	return m, nil
}

func (r *mockTeamMemberRepo) ListByUser(_ context.Context, userID string) ([]model.TeamMember, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var members []model.TeamMember
	for _, m := range r.members {
		if m.UserID == userID {
			members = append(members, *m)
		}
	}
	return members, nil
}

type mockTaskRepo struct {
	mu    sync.Mutex
	tasks map[string]*model.Task
}

func newMockTaskRepo() *mockTaskRepo {
	return &mockTaskRepo{tasks: make(map[string]*model.Task)}
}

func (r *mockTaskRepo) Create(_ context.Context, task *model.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[task.ID] = task
	return nil
}

func (r *mockTaskRepo) GetByID(_ context.Context, id string) (*model.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tasks[id]
	if !ok {
		return nil, nil
	}
	return t, nil
}

func (r *mockTaskRepo) List(_ context.Context, filter repository.TaskFilter) ([]model.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var tasks []model.Task
	for _, t := range r.tasks {
		if len(filter.TeamIDs) > 0 {
			found := false
			for _, id := range filter.TeamIDs {
				if t.TeamID == id {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		tasks = append(tasks, *t)
	}
	return tasks, nil
}

func (r *mockTaskRepo) Update(_ context.Context, task *model.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[task.ID] = task
	return nil
}

func (r *mockTaskRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tasks, id)
	return nil
}

type mockTaskHistoryRepo struct {
	mu      sync.Mutex
	entries []model.TaskHistory
}

func newMockTaskHistoryRepo() *mockTaskHistoryRepo {
	return &mockTaskHistoryRepo{}
}

func (r *mockTaskHistoryRepo) Create(_ context.Context, h *model.TaskHistory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, *h)
	return nil
}

func (r *mockTaskHistoryRepo) ListByTask(_ context.Context, taskID string) ([]model.TaskHistory, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []model.TaskHistory
	for _, e := range r.entries {
		if e.TaskID == taskID {
			result = append(result, e)
		}
	}
	return result, nil
}

type mockTaskCommentRepo struct {
	mu       sync.Mutex
	comments map[string]*model.TaskComment
}

func newMockTaskCommentRepo() *mockTaskCommentRepo {
	return &mockTaskCommentRepo{comments: make(map[string]*model.TaskComment)}
}

func (r *mockTaskCommentRepo) Create(_ context.Context, c *model.TaskComment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.comments[c.ID] = c
	return nil
}

func (r *mockTaskCommentRepo) GetByID(_ context.Context, taskID, commentID string) (*model.TaskComment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.comments[commentID]
	if !ok || c.TaskID != taskID {
		return nil, nil
	}
	return c, nil
}

func (r *mockTaskCommentRepo) ListByTask(_ context.Context, taskID string) ([]model.TaskComment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []model.TaskComment
	for _, c := range r.comments {
		if c.TaskID == taskID {
			result = append(result, *c)
		}
	}
	return result, nil
}

func (r *mockTaskCommentRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.comments, id)
	return nil
}

type mockDupKeyError struct{}

func (e *mockDupKeyError) Error() string { return "duplicate key" }
