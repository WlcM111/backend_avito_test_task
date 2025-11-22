package service

import (
	"context"

	"pr-reviewer-service/internal/domain"
)

// TeamService содержит бизнес-логику, связанную с командами.
type TeamService struct {
	teamRepo domain.TeamRepository
	userRepo domain.UserRepository
}

// NewTeamService создаёт новый TeamService.
func NewTeamService(teamRepo domain.TeamRepository, userRepo domain.UserRepository) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

// CreateTeam создаёт команду и добавляет/обновляет её участников.
func (s *TeamService) CreateTeam(ctx context.Context, teamName string, members []domain.User) (domain.Team, error) {
	exists, err := s.teamRepo.TeamExists(ctx, teamName)

	if err != nil {
		return domain.Team{}, err
	}

	if exists {
		return domain.Team{}, domain.NewDomainError(domain.ErrorCodeTeamExists, domain.ErrTeamExists)
	}

	if err := s.teamRepo.CreateTeam(ctx, teamName, members); err != nil {
		return domain.Team{}, err
	}

	// назначаем team_name пользователям и создаём/обновляем их
	if err := s.userRepo.UpsertUsers(ctx, teamName, members); err != nil {
		return domain.Team{}, err
	}

	team, err := s.teamRepo.GetTeamWithMembers(ctx, teamName)

	if err != nil {
		return domain.Team{}, err
	}

	return team, nil
}

// GetTeam возвращает информацию о команде по имени.
func (s *TeamService) GetTeam(ctx context.Context, teamName string) (domain.Team, error) {
	team, err := s.teamRepo.GetTeamWithMembers(ctx, teamName)

	if err != nil {
		if err == domain.ErrNotFound {
			return domain.Team{}, domain.NewDomainError(domain.ErrorCodeNotFound, err)
		}

		return domain.Team{}, err
	}

	return team, nil
}
