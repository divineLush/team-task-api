package service

import (
	"github.com/team-task-api/internal/repository"
	"github.com/team-task-api/pkg/database"
)

func newTxUserRepo(tx database.Querier) UserRepository {
	return repository.NewUserRepository(tx)
}

func newTxTeamRepo(tx database.Querier) TeamRepository {
	return repository.NewTeamRepository(tx)
}

func newTxTeamMemberRepo(tx database.Querier) TeamMemberRepository {
	return repository.NewTeamMemberRepository(tx)
}

func newTxTaskRepo(tx database.Querier) TaskRepository {
	return repository.NewTaskRepository(tx)
}

func newTxTaskHistoryRepo(tx database.Querier) TaskHistoryRepository {
	return repository.NewTaskHistoryRepository(tx)
}

func newTxTaskCommentRepo(tx database.Querier) TaskCommentRepository {
	return repository.NewTaskCommentRepository(tx)
}
