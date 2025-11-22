package domain

import "time"

// User описывает пользователя и его состояние в системе.
type User struct {
	ID        string
	Username  string
	TeamName  string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Team представляет команду и её участников.
type Team struct {
	Name    string
	Members []User
}

// PRStatus — статус pull request.
type PRStatus string

// Статусы pull request.
const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

// PullRequest описывает pull request с назначенными ревьюерами.
type PullRequest struct {
	ID                string
	Name              string
	AuthorID          string
	Status            PRStatus
	AssignedReviewers []string
	CreatedAt         *time.Time
	MergedAt          *time.Time
}

// PullRequestShort — краткая информация о pull request.
type PullRequestShort struct {
	ID       string
	Name     string
	AuthorID string
	Status   PRStatus
}

// AssignmentStatByUser содержит статистику назначений по пользователю.
type AssignmentStatByUser struct {
	UserID string
	Count  int64
}
