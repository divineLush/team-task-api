package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/team-task-api/internal/model"
	"github.com/team-task-api/pkg/database"
)

var (
	ErrNotMember     = errors.New("user is not a member of this team")
	ErrForbidden     = errors.New("insufficient permissions")
	ErrAlreadyMember = errors.New("user is already a member of this team")
	ErrTeamNotFound  = errors.New("team not found")
	ErrInvalidRole   = errors.New("owner role cannot be assigned via invite")
)

type TeamService struct {
	txm               database.TxManager
	teamRepo          TeamRepository
	teamMemberRepo    TeamMemberRepository
	newTeamRepoInTx   func(database.Querier) TeamRepository
	newMemberRepoInTx func(database.Querier) TeamMemberRepository
}

func NewTeamService(txm database.TxManager, teamRepo TeamRepository, teamMemberRepo TeamMemberRepository) *TeamService {
	return &TeamService{
		txm:               txm,
		teamRepo:          teamRepo,
		teamMemberRepo:    teamMemberRepo,
		newTeamRepoInTx:   newTxTeamRepo,
		newMemberRepoInTx: newTxTeamMemberRepo,
	}
}

func (s *TeamService) Create(ctx context.Context, userID string, req *model.CreateTeamRequest) (*model.Team, error) {
	var team *model.Team

	err := s.txm.InTx(ctx, func(tx database.Querier) error {
		txTeamRepo := s.newTeamRepoInTx(tx)
		txMemberRepo := s.newMemberRepoInTx(tx)

		team = &model.Team{
			Name:      req.Name,
			CreatedBy: userID,
		}

		if err := txTeamRepo.Create(ctx, team); err != nil {
			return fmt.Errorf("create team: %w", err)
		}

		member := &model.TeamMember{
			TeamID: team.ID,
			UserID: userID,
			Role:   model.RoleOwner,
		}
		if err := txMemberRepo.Add(ctx, member); err != nil {
			return fmt.Errorf("add owner as member: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
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
	return s.txm.InTx(ctx, func(tx database.Querier) error {
		txMemberRepo := s.newMemberRepoInTx(tx)

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
		if role == model.RoleOwner {
			return ErrInvalidRole
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

		return nil
	})
}

func (s *TeamService) GetByID(ctx context.Context, userID, teamID string) (*model.Team, error) {
	var team *model.Team

	err := s.txm.InTx(ctx, func(tx database.Querier) error {
		txTeamRepo := s.newTeamRepoInTx(tx)
		txMemberRepo := s.newMemberRepoInTx(tx)

		var err error
		team, err = txTeamRepo.GetByID(ctx, teamID)
		if err != nil {
			return fmt.Errorf("get team: %w", err)
		}
		if team == nil {
			return ErrTeamNotFound
		}

		member, err := txMemberRepo.Get(ctx, teamID, userID)
		if err != nil {
			return fmt.Errorf("check membership: %w", err)
		}
		if member == nil {
			return ErrNotMember
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (s *TeamService) Update(ctx context.Context, userID, teamID string, req *model.UpdateTeamRequest) (*model.Team, error) {
	var team *model.Team

	err := s.txm.InTx(ctx, func(tx database.Querier) error {
		txTeamRepo := s.newTeamRepoInTx(tx)
		txMemberRepo := s.newMemberRepoInTx(tx)

		var err error
		team, err = txTeamRepo.GetByID(ctx, teamID)
		if err != nil {
			return fmt.Errorf("get team: %w", err)
		}
		if team == nil {
			return ErrTeamNotFound
		}

		caller, err := txMemberRepo.Get(ctx, teamID, userID)
		if err != nil {
			return fmt.Errorf("check membership: %w", err)
		}
		if caller == nil {
			return ErrNotMember
		}
		if caller.Role != model.RoleOwner && caller.Role != model.RoleAdmin {
			return ErrForbidden
		}

		if req.Name != nil {
			team.Name = *req.Name
		}

		if err := txTeamRepo.Update(ctx, team); err != nil {
			return fmt.Errorf("update team: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (s *TeamService) Delete(ctx context.Context, userID, teamID string) error {
	return s.txm.InTx(ctx, func(tx database.Querier) error {
		txTeamRepo := s.newTeamRepoInTx(tx)
		txMemberRepo := s.newMemberRepoInTx(tx)

		team, err := txTeamRepo.GetByID(ctx, teamID)
		if err != nil {
			return fmt.Errorf("get team: %w", err)
		}
		if team == nil {
			return ErrTeamNotFound
		}

		caller, err := txMemberRepo.Get(ctx, teamID, userID)
		if err != nil {
			return fmt.Errorf("check membership: %w", err)
		}
		if caller == nil {
			return ErrNotMember
		}
		if caller.Role != model.RoleOwner {
			return ErrForbidden
		}

		if err := txTeamRepo.Delete(ctx, teamID); err != nil {
			return fmt.Errorf("delete team: %w", err)
		}

		return nil
	})
}
