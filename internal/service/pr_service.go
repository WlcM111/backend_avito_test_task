package service

import (
	"context"
	"time"

	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/internal/random"
)

// PullRequestService содержит бизнес-логику, связанную с pull request-ами.
type PullRequestService struct {
	prRepo   domain.PullRequestRepository
	userRepo domain.UserRepository
	teamRepo domain.TeamRepository
	rand     random.Rand
}

// NewPullRequestService создаёт новый PullRequestService.
func NewPullRequestService(
	prRepo domain.PullRequestRepository,
	userRepo domain.UserRepository,
	teamRepo domain.TeamRepository,
	rand random.Rand,
) *PullRequestService {
	return &PullRequestService{
		prRepo:   prRepo,
		userRepo: userRepo,
		teamRepo: teamRepo,
		rand:     rand,
	}
}

// CreatePR создаёт pull request и автоматически назначает ревьюеров.
func (s *PullRequestService) CreatePR(ctx context.Context, id, name, authorID string) (domain.PullRequest, error) {
	exists, err := s.prRepo.PRExists(ctx, id)

	if err != nil {
		return domain.PullRequest{}, err
	}

	if exists {
		return domain.PullRequest{}, domain.NewDomainError(domain.ErrorCodePRExists, domain.ErrPRExists)
	}

	author, err := s.userRepo.GetByID(ctx, authorID)

	if err != nil {
		if err == domain.ErrNotFound {
			return domain.PullRequest{}, domain.NewDomainError(domain.ErrorCodeNotFound, err)
		}

		return domain.PullRequest{}, err
	}

	teamName := author.TeamName

	if teamName == "" {
		return domain.PullRequest{}, domain.NewDomainError(domain.ErrorCodeNotFound, domain.ErrNotFound)
	}

	candidates, err := s.userRepo.GetActiveTeamMembersExcept(ctx, teamName, authorID)

	if err != nil {
		return domain.PullRequest{}, err
	}

	assigned := chooseReviewers(candidates, 2, s.rand)

	now := time.Now().UTC()
	pr := domain.PullRequest{
		ID:                id,
		Name:              name,
		AuthorID:          authorID,
		Status:            domain.PRStatusOpen,
		AssignedReviewers: assigned,
		CreatedAt:         &now,
		MergedAt:          nil,
	}

	if err := s.prRepo.Create(ctx, pr); err != nil {
		return domain.PullRequest{}, err
	}

	created, err := s.prRepo.GetByID(ctx, id)

	if err != nil {
		return domain.PullRequest{}, err
	}

	return created, nil
}

func chooseReviewers(users []domain.User, max int, rand random.Rand) []string {
	if len(users) == 0 || max <= 0 {
		return nil
	}

	// копируем, чтобы не мутировать исходный слайс
	tmp := make([]domain.User, len(users))
	copy(tmp, users)

	// Fisher–Yates
	for i := len(tmp) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		tmp[i], tmp[j] = tmp[j], tmp[i]
	}

	if len(tmp) > max {
		tmp = tmp[:max]
	}

	result := make([]string, 0, len(tmp))

	for _, u := range tmp {
		result = append(result, u.ID)
	}

	return result
}

// MergePR помечает pull request как merged (идемпотентно).
func (s *PullRequestService) MergePR(ctx context.Context, id string) (domain.PullRequest, error) {
	pr, err := s.prRepo.GetByID(ctx, id)

	if err != nil {
		if err == domain.ErrNotFound {
			return domain.PullRequest{}, domain.NewDomainError(domain.ErrorCodeNotFound, err)
		}

		return domain.PullRequest{}, err
	}

	if pr.Status == domain.PRStatusMerged {
		return pr, nil
	}

	now := time.Now().UTC()
	merged, err := s.prRepo.MarkMerged(ctx, id, now)

	if err != nil {
		if err == domain.ErrNotFound {
			return domain.PullRequest{}, domain.NewDomainError(domain.ErrorCodeNotFound, err)
		}

		return domain.PullRequest{}, err
	}

	return merged, nil
}

// ReassignReviewer переназначает ревьюера в pull request на другого активного участника команды.
// nolint:gocyclo
func (s *PullRequestService) ReassignReviewer(
	ctx context.Context,
	prID, oldReviewerID string,
) (pr domain.PullRequest, replacedBy string, err error) {
	pr, err = s.prRepo.GetByID(ctx, prID)

	if err != nil {
		if err == domain.ErrNotFound {
			err = domain.NewDomainError(domain.ErrorCodeNotFound, err)
		}

		return
	}

	// ensure user exists
	if _, uerr := s.userRepo.GetByID(ctx, oldReviewerID); uerr != nil {
		if uerr == domain.ErrNotFound {
			err = domain.NewDomainError(domain.ErrorCodeNotFound, uerr)
			return
		}

		err = uerr
		return
	}

	if pr.Status == domain.PRStatusMerged {
		err = domain.NewDomainError(domain.ErrorCodePRMerged, domain.ErrPRMerged)
		return
	}

	isAssigned := false
	assignedSet := make(map[string]struct{}, len(pr.AssignedReviewers))

	for _, id := range pr.AssignedReviewers {
		assignedSet[id] = struct{}{}

		if id == oldReviewerID {
			isAssigned = true
		}
	}

	if !isAssigned {
		err = domain.NewDomainError(domain.ErrorCodeNotAssigned, domain.ErrReviewerNotAssigned)
		return
	}

	teamName, err := s.userRepo.GetTeamByUserID(ctx, oldReviewerID)

	if err != nil {
		if err == domain.ErrNotFound {
			err = domain.NewDomainError(domain.ErrorCodeNotFound, err)
		}

		return
	}

	candidates, err := s.userRepo.GetActiveTeamMembersExcept(ctx, teamName, oldReviewerID)
	if err != nil {
		return
	}

	// исключаем уже назначенных ревьюверов
	filtered := make([]domain.User, 0, len(candidates))

	for _, c := range candidates {
		if _, ok := assignedSet[c.ID]; ok {
			continue
		}

		filtered = append(filtered, c)
	}

	if len(filtered) == 0 {
		err = domain.NewDomainError(domain.ErrorCodeNoCandidate, domain.ErrNoCandidate)
		return
	}

	// выбираем случайного кандидата
	idx := s.rand.Intn(len(filtered))
	newReviewer := filtered[idx].ID

	updated, err := s.prRepo.ReassignReviewer(ctx, prID, oldReviewerID, newReviewer)

	if err != nil {
		if err == domain.ErrReviewerNotAssigned {
			err = domain.NewDomainError(domain.ErrorCodeNotAssigned, err)
		}

		return
	}

	return updated, newReviewer, nil
}
