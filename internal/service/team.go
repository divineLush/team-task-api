package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/team-task-api/internal/model"
	"github.com/team-task-api/internal/repository"
)

var (
	ErrNotMember     = errors.New("user is not a member of this team")
	ErrForbidden     = errors.New("insufficient permissions")
	ErrAlreadyMember = errors.New("user is already a member of this team")
)

type TeamService struct {
	db             *sql.DB
	teamRepo       *repository.TeamRepository
	teamMemberRepo *repository.TeamMemberRepository
}

func NewTeamService(db *sql.DB, teamRepo *repository.TeamRepository, teamMemberRepo *repository.TeamMemberRepository) *TeamService {
	return &TeamService{db: db, teamRepo: teamRepo, teamMemberRepo: teamMemberRepo}
}

func (s *TeamService) Create(ctx context.Context, userID string, req *model.CreateTeamRequest) (*model.Team, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	txTeamRepo := repository.NewTeamRepository(tx)
	txMemberRepo := repository.NewTeamMemberRepository(tx)

	team := &model.Team{
		Name:      req.Name,
		CreatedBy: userID,
	}

	if err := txTeamRepo.Create(ctx, team); err != nil {
		return nil, fmt.Errorf("create team: %w", err)
	}

	member := &model.TeamMember{
		TeamID: team.ID,
		UserID: userID,
		Role:   model.RoleOwner,
	}
	if err := txMemberRepo.Add(ctx, member); err != nil {
		return nil, fmt.Errorf("add owner as member: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return team, nil
}

func (s *TeamService) ListByUser(ctx context.Context, userID string) ([]model.Team, error) {
	teams, err := s.teamRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list teams by user: %w", err)
	}
	return teams, nil
}

func (s *TeamService) Invite(ctx context.Context, callerID, teamID string, req *model.InviteRequest) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	txMemberRepo := repository.NewTeamMemberRepository(tx)

	caller, err := txMemberRepo.Get(ctx, teamID, callerID)
	if err != nil {
		return fmt.Errorf("get caller membership: %w", err)
	}
	if caller == nil {
		return ErrNotMember
	}
	if caller.Role != model.RoleOwner && caller.Role != model.RoleAdmin {
		return ErrForbidden
	}

	existing, err := txMemberRepo.Get(ctx, teamID, req.UserID)
	if err != nil {
		return fmt.Errorf("check existing membership: %w", err)
	}
	if existing != nil {
		return ErrAlreadyMember
	}

	role := req.Role
	if role == "" {
		role = model.RoleMember
	}

	member := &model.TeamMember{
		TeamID: teamID,
		UserID: req.UserID,
		Role:   role,
	}
	if err := txMemberRepo.Add(ctx, member); err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrAlreadyMember
		}
		return fmt.Errorf("add team member: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
