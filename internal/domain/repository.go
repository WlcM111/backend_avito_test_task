package domain

import (
	"context"
	"database/sql"
	"time"
)

// TeamRepository описывает операции работы с командами.
type TeamRepository interface {
	CreateTeam(ctx context.Context, name string, members []User) error
	GetTeamWithMembers(ctx context.Context, teamName string) (Team, error)
	TeamExists(ctx context.Context, name string) (bool, error)
}

// UserRepository описывает операции работы с пользователями.
type UserRepository interface {
	GetByID(ctx context.Context, id string) (User, error)
	UpsertUsers(ctx context.Context, teamName string, users []User) error
	SetIsActive(ctx context.Context, id string, isActive bool) (User, error)
	GetActiveTeamMembersExcept(ctx context.Context, teamName, excludeUserID string) ([]User, error)
	GetTeamByUserID(ctx context.Context, userID string) (string, error)
}

// PullRequestRepository описывает операции с pull request-ами.
type PullRequestRepository interface {
	Create(ctx context.Context, pr PullRequest) error
	GetByID(ctx context.Context, id string) (PullRequest, error)
	MarkMerged(ctx context.Context, id string, mergedAt time.Time) (PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) (PullRequest, error)
	ListByReviewer(ctx context.Context, reviewerID string) ([]PullRequestShort, error)
	PRExists(ctx context.Context, id string) (bool, error)
	GetAssignmentStatsByUser(ctx context.Context) ([]AssignmentStatByUser, error)
	WithTx(ctx context.Context, fn func(ctx context.Context, tx *sql.Tx) error) error
}
