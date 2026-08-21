package service

import (
	"context"
	"testing"

	"github.com/team-task-api/internal/model"
	"github.com/team-task-api/pkg/database"
)

func overrideTeamFactories(svc *TeamService, tr TeamRepository, mr TeamMemberRepository) {
	svc.newTeamRepoInTx = func(_ database.Querier) TeamRepository { return tr }
	svc.newMemberRepoInTx = func(_ database.Querier) TeamMemberRepository { return mr }
}

func TestTeamCreate_OwnerIsMember(t *testing.T) {
	txm := &mockTxManager{}
	teamRepo := newMockTeamRepo()
	memberRepo := newMockTeamMemberRepo()
	svc := NewTeamService(txm, teamRepo, memberRepo)
	overrideTeamFactories(svc, teamRepo, memberRepo)

	team, err := svc.Create(context.Background(), "user-1", &model.CreateTeamRequest{Name: "My Team"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if team.Name != "My Team" {
		t.Errorf("expected name My Team, got %s", team.Name)
	}

	member, err := memberRepo.Get(context.Background(), team.ID, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if member == nil {
		t.Fatal("expected creator to be a team member")
	}
	if member.Role != model.RoleOwner {
		t.Errorf("expected role owner, got %s", member.Role)
	}
}

func TestTeamGetByID_MemberCanAccess(t *testing.T) {
	txm := &mockTxManager{}
	teamRepo := newMockTeamRepo()
	memberRepo := newMockTeamMemberRepo()
	svc := NewTeamService(txm, teamRepo, memberRepo)
	overrideTeamFactories(svc, teamRepo, memberRepo)

	team, _ := svc.Create(context.Background(), "owner", &model.CreateTeamRequest{Name: "Team"})
	memberRepo.Add(context.Background(), &model.TeamMember{
		TeamID: team.ID, UserID: "member-1", Role: model.RoleMember,
	})

	got, err := svc.GetByID(context.Background(), "member-1", team.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != team.ID {
		t.Errorf("expected team ID %s, got %s", team.ID, got.ID)
	}
}

func TestTeamGetByID_NonMemberCannotAccess(t *testing.T) {
	txm := &mockTxManager{}
	teamRepo := newMockTeamRepo()
	memberRepo := newMockTeamMemberRepo()
	svc := NewTeamService(txm, teamRepo, memberRepo)
	overrideTeamFactories(svc, teamRepo, memberRepo)

	team, _ := svc.Create(context.Background(), "owner", &model.CreateTeamRequest{Name: "Team"})

	_, err := svc.GetByID(context.Background(), "outsider", team.ID)
	if err != ErrNotTeamMember {
		t.Fatalf("expected ErrNotTeamMember, got %v", err)
	}
}

func TestTeamUpdate_OwnerCanUpdate(t *testing.T) {
	txm := &mockTxManager{}
	teamRepo := newMockTeamRepo()
	memberRepo := newMockTeamMemberRepo()
	svc := NewTeamService(txm, teamRepo, memberRepo)
	overrideTeamFactories(svc, teamRepo, memberRepo)

	team, _ := svc.Create(context.Background(), "owner", &model.CreateTeamRequest{Name: "Old Name"})

	updated, err := svc.Update(context.Background(), "owner", team.ID, &model.UpdateTeamRequest{Name: strPtr("New Name")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "New Name" {
		t.Errorf("expected name New Name, got %s", updated.Name)
	}
}

func TestTeamUpdate_AdminCanUpdate(t *testing.T) {
	txm := &mockTxManager{}
	teamRepo := newMockTeamRepo()
	memberRepo := newMockTeamMemberRepo()
	svc := NewTeamService(txm, teamRepo, memberRepo)
	overrideTeamFactories(svc, teamRepo, memberRepo)

	team, _ := svc.Create(context.Background(), "owner", &model.CreateTeamRequest{Name: "Old"})
	memberRepo.Add(context.Background(), &model.TeamMember{
		TeamID: team.ID, UserID: "admin-1", Role: model.RoleAdmin,
	})

	updated, err := svc.Update(context.Background(), "admin-1", team.ID, &model.UpdateTeamRequest{Name: strPtr("Updated")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "Updated" {
		t.Errorf("expected name Updated, got %s", updated.Name)
	}
}

func TestTeamUpdate_MemberCannotUpdate(t *testing.T) {
	txm := &mockTxManager{}
	teamRepo := newMockTeamRepo()
	memberRepo := newMockTeamMemberRepo()
	svc := NewTeamService(txm, teamRepo, memberRepo)
	overrideTeamFactories(svc, teamRepo, memberRepo)

	team, _ := svc.Create(context.Background(), "owner", &model.CreateTeamRequest{Name: "Original"})
	memberRepo.Add(context.Background(), &model.TeamMember{
		TeamID: team.ID, UserID: "member-1", Role: model.RoleMember,
	})

	_, err := svc.Update(context.Background(), "member-1", team.ID, &model.UpdateTeamRequest{Name: strPtr("Hacked")})
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestTeamDelete_OwnerCanDelete(t *testing.T) {
	txm := &mockTxManager{}
	teamRepo := newMockTeamRepo()
	memberRepo := newMockTeamMemberRepo()
	svc := NewTeamService(txm, teamRepo, memberRepo)
	overrideTeamFactories(svc, teamRepo, memberRepo)

	team, _ := svc.Create(context.Background(), "owner", &model.CreateTeamRequest{Name: "Doomed"})

	err := svc.Delete(context.Background(), "owner", team.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTeamDelete_NonOwnerCannotDelete(t *testing.T) {
	txm := &mockTxManager{}
	teamRepo := newMockTeamRepo()
	memberRepo := newMockTeamMemberRepo()
	svc := NewTeamService(txm, teamRepo, memberRepo)
	overrideTeamFactories(svc, teamRepo, memberRepo)

	team, _ := svc.Create(context.Background(), "owner", &model.CreateTeamRequest{Name: "Safe"})
	memberRepo.Add(context.Background(), &model.TeamMember{
		TeamID: team.ID, UserID: "admin-1", Role: model.RoleAdmin,
	})

	err := svc.Delete(context.Background(), "admin-1", team.ID)
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestTeamInvite_OwnerCanInvite(t *testing.T) {
	txm := &mockTxManager{}
	teamRepo := newMockTeamRepo()
	memberRepo := newMockTeamMemberRepo()
	svc := NewTeamService(txm, teamRepo, memberRepo)
	overrideTeamFactories(svc, teamRepo, memberRepo)

	team, _ := svc.Create(context.Background(), "owner", &model.CreateTeamRequest{Name: "Team"})

	err := svc.Invite(context.Background(), "owner", team.ID, &model.InviteRequest{
		UserID: "new-member",
		Role:   model.RoleMember,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	member, _ := memberRepo.Get(context.Background(), team.ID, "new-member")
	if member == nil {
		t.Fatal("expected new member to be added")
	}
}

func TestTeamInvite_MemberCannotInvite(t *testing.T) {
	txm := &mockTxManager{}
	teamRepo := newMockTeamRepo()
	memberRepo := newMockTeamMemberRepo()
	svc := NewTeamService(txm, teamRepo, memberRepo)
	overrideTeamFactories(svc, teamRepo, memberRepo)

	team, _ := svc.Create(context.Background(), "owner", &model.CreateTeamRequest{Name: "Team"})
	memberRepo.Add(context.Background(), &model.TeamMember{
		TeamID: team.ID, UserID: "member-1", Role: model.RoleMember,
	})

	err := svc.Invite(context.Background(), "member-1", team.ID, &model.InviteRequest{
		UserID: "new-member",
	})
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestTeamInvite_CannotAssignOwnerRole(t *testing.T) {
	txm := &mockTxManager{}
	teamRepo := newMockTeamRepo()
	memberRepo := newMockTeamMemberRepo()
	svc := NewTeamService(txm, teamRepo, memberRepo)
	overrideTeamFactories(svc, teamRepo, memberRepo)

	team, _ := svc.Create(context.Background(), "owner", &model.CreateTeamRequest{Name: "Team"})

	err := svc.Invite(context.Background(), "owner", team.ID, &model.InviteRequest{
		UserID: "new-member",
		Role:   model.RoleOwner,
	})
	if err != ErrInvalidRole {
		t.Fatalf("expected ErrInvalidRole, got %v", err)
	}
}

func TestTeamInvite_AlreadyMember(t *testing.T) {
	txm := &mockTxManager{}
	teamRepo := newMockTeamRepo()
	memberRepo := newMockTeamMemberRepo()
	svc := NewTeamService(txm, teamRepo, memberRepo)
	overrideTeamFactories(svc, teamRepo, memberRepo)

	team, _ := svc.Create(context.Background(), "owner", &model.CreateTeamRequest{Name: "Team"})

	err := svc.Invite(context.Background(), "owner", team.ID, &model.InviteRequest{
		UserID: "already-here",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = svc.Invite(context.Background(), "owner", team.ID, &model.InviteRequest{
		UserID: "already-here",
	})
	if err != ErrAlreadyMember {
		t.Fatalf("expected ErrAlreadyMember, got %v", err)
	}
}

func TestTeamStats_OwnerCanView(t *testing.T) {
	txm := &mockTxManager{}
	teamRepo := newMockTeamRepo()
	memberRepo := newMockTeamMemberRepo()
	svc := NewTeamService(txm, teamRepo, memberRepo)
	overrideTeamFactories(svc, teamRepo, memberRepo)

	team, _ := svc.Create(context.Background(), "owner", &model.CreateTeamRequest{Name: "Team"})

	stats, err := svc.Stats(context.Background(), "owner", team.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats == nil {
		t.Fatal("expected stats, got nil")
	}
}

func TestTeamStats_AdminCanView(t *testing.T) {
	txm := &mockTxManager{}
	teamRepo := newMockTeamRepo()
	memberRepo := newMockTeamMemberRepo()
	svc := NewTeamService(txm, teamRepo, memberRepo)
	overrideTeamFactories(svc, teamRepo, memberRepo)

	team, _ := svc.Create(context.Background(), "owner", &model.CreateTeamRequest{Name: "Team"})
	memberRepo.Add(context.Background(), &model.TeamMember{
		TeamID: team.ID, UserID: "admin-1", Role: model.RoleAdmin,
	})

	stats, err := svc.Stats(context.Background(), "admin-1", team.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats == nil {
		t.Fatal("expected stats, got nil")
	}
}

func TestTeamStats_MemberCannotView(t *testing.T) {
	txm := &mockTxManager{}
	teamRepo := newMockTeamRepo()
	memberRepo := newMockTeamMemberRepo()
	svc := NewTeamService(txm, teamRepo, memberRepo)
	overrideTeamFactories(svc, teamRepo, memberRepo)

	team, _ := svc.Create(context.Background(), "owner", &model.CreateTeamRequest{Name: "Team"})
	memberRepo.Add(context.Background(), &model.TeamMember{
		TeamID: team.ID, UserID: "member-1", Role: model.RoleMember,
	})

	_, err := svc.Stats(context.Background(), "member-1", team.ID)
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestTeamStats_NonMemberCannotView(t *testing.T) {
	txm := &mockTxManager{}
	teamRepo := newMockTeamRepo()
	memberRepo := newMockTeamMemberRepo()
	svc := NewTeamService(txm, teamRepo, memberRepo)
	overrideTeamFactories(svc, teamRepo, memberRepo)

	team, _ := svc.Create(context.Background(), "owner", &model.CreateTeamRequest{Name: "Team"})

	_, err := svc.Stats(context.Background(), "outsider", team.ID)
	if err != ErrNotTeamMember {
		t.Fatalf("expected ErrNotTeamMember, got %v", err)
	}
}

func TestTeamStats_TeamNotFound(t *testing.T) {
	txm := &mockTxManager{}
	teamRepo := newMockTeamRepo()
	memberRepo := newMockTeamMemberRepo()
	svc := NewTeamService(txm, teamRepo, memberRepo)
	overrideTeamFactories(svc, teamRepo, memberRepo)

	_, err := svc.Stats(context.Background(), "owner", "nonexistent")
	if err != ErrTeamNotFound {
		t.Fatalf("expected ErrTeamNotFound, got %v", err)
	}
}
