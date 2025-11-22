package service

import (
	"context"

	"pr-reviewer-service/internal/domain"
)

// StatsService содержит бизнес-логику, связанную со статистикой по ревью.
type StatsService struct {
	prRepo domain.PullRequestRepository
}

// NewStatsService создаёт новый StatsService.
func NewStatsService(prRepo domain.PullRequestRepository) *StatsService {
	return &StatsService{
		prRepo: prRepo,
	}
}

// GetAssignmentsByUser возвращает статистику назначений на ревью по пользователям.
func (s *StatsService) GetAssignmentsByUser(ctx context.Context) ([]domain.AssignmentStatByUser, error) {
	return s.prRepo.GetAssignmentStatsByUser(ctx)
}
