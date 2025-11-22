package service

import (
	"context"

	"pr-reviewer-service/internal/domain"
)

// UserService содержит бизнес-логику, связанную с пользователями.
type UserService struct {
	userRepo domain.UserRepository
	prRepo   domain.PullRequestRepository
}

// NewUserService создаёт новый UserService.
func NewUserService(userRepo domain.UserRepository, prRepo domain.PullRequestRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
		prRepo:   prRepo,
	}
}

// SetIsActive изменяет флаг активности пользователя и возвращает обновлённую сущность.
func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) (domain.User, error) {
	user, err := s.userRepo.SetIsActive(ctx, userID, isActive)

	if err != nil {
		if err == domain.ErrNotFound {
			return domain.User{}, domain.NewDomainError(domain.ErrorCodeNotFound, err)
		}

		return domain.User{}, err
	}

	return user, nil
}

// GetReviewPRs возвращает список PR для ревью указанного пользователя.
func (s *UserService) GetReviewPRs(ctx context.Context, userID string) (string, []domain.PullRequestShort, error) {
	user, err := s.userRepo.GetByID(ctx, userID)

	if err != nil {
		if err == domain.ErrNotFound {
			return "", nil, domain.NewDomainError(domain.ErrorCodeNotFound, err)
		}

		return "", nil, err
	}

	prs, err := s.prRepo.ListByReviewer(ctx, userID)

	if err != nil {
		return "", nil, err
	}

	return user.ID, prs, nil
}
