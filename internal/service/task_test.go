package service

import (
	"context"
	"testing"

	"github.com/team-task-api/internal/model"
	"github.com/team-task-api/internal/repository"
	"github.com/team-task-api/pkg/database"
)

func strP(s string) *string { return &s }

func setupTaskService() (*TaskService, *mockTaskRepo, *mockTeamMemberRepo, *mockTaskHistoryRepo) {
	txm := &mockTxManager{}
	taskRepo := newMockTaskRepo()
	memberRepo := newMockTeamMemberRepo()
	historyRepo := newMockTaskHistoryRepo()
	svc := NewTaskService(txm, taskRepo, memberRepo, nil, nil)
	svc.newTaskRepoInTx = func(_ database.Querier) TaskRepository { return taskRepo }
	svc.newMemberRepoInTx = func(_ database.Querier) TeamMemberRepository { return memberRepo }
	svc.newHistoryRepoInTx = func(_ database.Querier) TaskHistoryRepository { return historyRepo }
	return svc, taskRepo, memberRepo, historyRepo
}

func seedTeamMember(memberRepo *mockTeamMemberRepo, teamID, userID string, role model.TeamRole) {
	memberRepo.Add(context.Background(), &model.TeamMember{
		TeamID: teamID, UserID: userID, Role: role,
	})
}

func seedTask(taskRepo *mockTaskRepo, task *model.Task) {
	taskRepo.Create(context.Background(), task)
}

func TestTaskCreate_MemberCanCreate(t *testing.T) {
	svc, _, memberRepo, _ := setupTaskService()
	seedTeamMember(memberRepo, "team-1", "user-1", model.RoleMember)

	task, err := svc.Create(context.Background(), "user-1", &model.CreateTaskRequest{
		TeamID: "team-1",
		Title:  "My Task",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Title != "My Task" {
		t.Errorf("expected title My Task, got %s", task.Title)
	}
	if task.Status != model.StatusPending {
		t.Errorf("expected status pending, got %s", task.Status)
	}
	if task.CreatedBy != "user-1" {
		t.Errorf("expected created_by user-1, got %s", task.CreatedBy)
	}
}

func TestTaskCreate_NonMemberCannotCreate(t *testing.T) {
	svc, _, _, _ := setupTaskService()

	_, err := svc.Create(context.Background(), "outsider", &model.CreateTaskRequest{
		TeamID: "team-1",
		Title:  "Hacked Task",
	})
	if err != ErrNotTeamMember {
		t.Fatalf("expected ErrNotTeamMember, got %v", err)
	}
}

func TestTaskGetByID_MemberCanAccess(t *testing.T) {
	svc, taskRepo, memberRepo, _ := setupTaskService()
	seedTeamMember(memberRepo, "team-1", "user-1", model.RoleMember)
	seedTask(taskRepo, &model.Task{ID: "task-1", TeamID: "team-1", Title: "T", CreatedBy: "user-1"})

	task, err := svc.GetByID(context.Background(), "user-1", "task-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.ID != "task-1" {
		t.Errorf("expected task ID task-1, got %s", task.ID)
	}
}

func TestTaskGetByID_NonMemberCannotAccess(t *testing.T) {
	svc, taskRepo, _, _ := setupTaskService()
	seedTask(taskRepo, &model.Task{ID: "task-1", TeamID: "team-1", Title: "T", CreatedBy: "user-1"})

	_, err := svc.GetByID(context.Background(), "outsider", "task-1")
	if err != ErrNotTeamMember {
		t.Fatalf("expected ErrNotTeamMember, got %v", err)
	}
}

func TestTaskUpdate_CreatorCanEdit(t *testing.T) {
	svc, taskRepo, memberRepo, historyRepo := setupTaskService()
	seedTeamMember(memberRepo, "team-1", "creator", model.RoleMember)
	seedTask(taskRepo, &model.Task{
		ID: "task-1", TeamID: "team-1", Title: "Old", CreatedBy: "creator", Status: model.StatusPending,
	})

	task, err := svc.Update(context.Background(), "creator", "task-1", &model.UpdateTaskRequest{
		Title: strP("New"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Title != "New" {
		t.Errorf("expected title New, got %s", task.Title)
	}
	if len(historyRepo.entries) != 1 {
		t.Errorf("expected 1 history entry, got %d", len(historyRepo.entries))
	}
}

func TestTaskUpdate_AssigneeCanEdit(t *testing.T) {
	svc, taskRepo, memberRepo, _ := setupTaskService()
	seedTeamMember(memberRepo, "team-1", "assignee", model.RoleMember)
	seedTask(taskRepo, &model.Task{
		ID: "task-1", TeamID: "team-1", Title: "Old", CreatedBy: "creator", AssigneeID: strP("assignee"), Status: model.StatusPending,
	})

	done := model.StatusDone
	_, err := svc.Update(context.Background(), "assignee", "task-1", &model.UpdateTaskRequest{
		Status: &done,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTaskUpdate_NonCreatorNonAssigneeCannotEdit(t *testing.T) {
	svc, taskRepo, memberRepo, _ := setupTaskService()
	seedTeamMember(memberRepo, "team-1", "member-1", model.RoleMember)
	seedTask(taskRepo, &model.Task{
		ID: "task-1", TeamID: "team-1", Title: "T", CreatedBy: "other", Status: model.StatusPending,
	})

	_, err := svc.Update(context.Background(), "member-1", "task-1", &model.UpdateTaskRequest{
		Title: strP("Hacked"),
	})
	if err != ErrNotAuthorized {
		t.Fatalf("expected ErrNotAuthorized, got %v", err)
	}
}

func TestTaskUpdate_OwnerCanEditAny(t *testing.T) {
	svc, taskRepo, memberRepo, _ := setupTaskService()
	seedTeamMember(memberRepo, "team-1", "owner", model.RoleOwner)
	seedTask(taskRepo, &model.Task{
		ID: "task-1", TeamID: "team-1", Title: "T", CreatedBy: "someone-else", Status: model.StatusPending,
	})

	task, err := svc.Update(context.Background(), "owner", "task-1", &model.UpdateTaskRequest{
		Title: strP("Owner Edit"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Title != "Owner Edit" {
		t.Errorf("expected title Owner Edit, got %s", task.Title)
	}
}

func TestTaskUpdate_AdminCanEditAny(t *testing.T) {
	svc, taskRepo, memberRepo, _ := setupTaskService()
	seedTeamMember(memberRepo, "team-1", "admin", model.RoleAdmin)
	seedTask(taskRepo, &model.Task{
		ID: "task-1", TeamID: "team-1", Title: "T", CreatedBy: "someone-else", Status: model.StatusPending,
	})

	task, err := svc.Update(context.Background(), "admin", "task-1", &model.UpdateTaskRequest{
		Title: strP("Admin Edit"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Title != "Admin Edit" {
		t.Errorf("expected title Admin Edit, got %s", task.Title)
	}
}

func TestTaskUpdate_CannotReassignIfNotPrivilegedOrCreator(t *testing.T) {
	svc, taskRepo, memberRepo, _ := setupTaskService()
	seedTeamMember(memberRepo, "team-1", "assignee", model.RoleMember)
	seedTeamMember(memberRepo, "team-1", "other", model.RoleMember)
	seedTask(taskRepo, &model.Task{
		ID: "task-1", TeamID: "team-1", Title: "T", CreatedBy: "creator", AssigneeID: strP("assignee"), Status: model.StatusPending,
	})

	_, err := svc.Update(context.Background(), "assignee", "task-1", &model.UpdateTaskRequest{
		AssigneeID: strP("other"),
	})
	if err != ErrCannotReassign {
		t.Fatalf("expected ErrCannotReassign, got %v", err)
	}
}

func TestTaskUpdate_AssigneeNotMember(t *testing.T) {
	svc, taskRepo, memberRepo, _ := setupTaskService()
	seedTeamMember(memberRepo, "team-1", "creator", model.RoleMember)
	seedTask(taskRepo, &model.Task{
		ID: "task-1", TeamID: "team-1", Title: "T", CreatedBy: "creator", Status: model.StatusPending,
	})

	_, err := svc.Update(context.Background(), "creator", "task-1", &model.UpdateTaskRequest{
		AssigneeID: strP("outsider"),
	})
	if err != ErrAssigneeNotMember {
		t.Fatalf("expected ErrAssigneeNotMember, got %v", err)
	}
}

func TestTaskDelete_MemberCanDelete(t *testing.T) {
	svc, taskRepo, memberRepo, _ := setupTaskService()
	seedTeamMember(memberRepo, "team-1", "user-1", model.RoleMember)
	seedTask(taskRepo, &model.Task{ID: "task-1", TeamID: "team-1", Title: "T", CreatedBy: "user-1"})

	err := svc.Delete(context.Background(), "user-1", "task-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTaskDelete_NonMemberCannotDelete(t *testing.T) {
	svc, taskRepo, _, _ := setupTaskService()
	seedTask(taskRepo, &model.Task{ID: "task-1", TeamID: "team-1", Title: "T", CreatedBy: "user-1"})

	err := svc.Delete(context.Background(), "outsider", "task-1")
	if err != ErrNotTeamMember {
		t.Fatalf("expected ErrNotTeamMember, got %v", err)
	}
}

func TestTaskList_EmptyForMemberWithNoTeams(t *testing.T) {
	svc, _, _, _ := setupTaskService()

	tasks, err := svc.List(context.Background(), "no-teams", repository.TaskFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks))
	}
}

func TestTaskList_OnlyShowsTeamsMemberBelongsTo(t *testing.T) {
	svc, taskRepo, memberRepo, _ := setupTaskService()
	seedTeamMember(memberRepo, "team-1", "user-1", model.RoleMember)
	seedTask(taskRepo, &model.Task{ID: "t1", TeamID: "team-1", Title: "My Task", CreatedBy: "user-1"})
	seedTask(taskRepo, &model.Task{ID: "t2", TeamID: "team-2", Title: "Other Task", CreatedBy: "other"})

	tasks, err := svc.List(context.Background(), "user-1", repository.TaskFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].ID != "t1" {
		t.Errorf("expected task t1, got %s", tasks[0].ID)
	}
}
