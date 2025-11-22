package httpapi

import "time"

// TeamMemberRequest описывает участника команды в запросе на создание команды.
type TeamMemberRequest struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

// TeamRequest — тело запроса на создание/обновление команды.
type TeamRequest struct {
	TeamName string              `json:"team_name"`
	Members  []TeamMemberRequest `json:"members"`
}

// TeamMemberDTO — участник команды в ответе API.
type TeamMemberDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

// TeamDTO — команда в ответах API.
type TeamDTO struct {
	TeamName string          `json:"team_name"`
	Members  []TeamMemberDTO `json:"members"`
}

// TeamCreateResponse — ответ API при создании команды.
type TeamCreateResponse struct {
	Team TeamDTO `json:"team"`
}

// SetIsActiveRequest — запрос на изменение активности пользователя.
type SetIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

// UserDTO — модель пользователя в HTTP-слое.
type UserDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

// SetIsActiveResponse — ответ API после изменения активности пользователя.
type SetIsActiveResponse struct {
	User UserDTO `json:"user"`
}

// CreatePRRequest — запрос на создание pull request.
type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

// PullRequestDTO — модель pull request в HTTP-слое.
type PullRequestDTO struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         *time.Time `json:"createdAt,omitempty"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}

// CreatePRResponse — ответ API после создания pull request.
type CreatePRResponse struct {
	PR PullRequestDTO `json:"pr"`
}

// MergePRRequest — запрос на пометку pull request как слитого (merged).
type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

// MergePRResponse — ответ API после успешного merge PR.
type MergePRResponse struct {
	PR PullRequestDTO `json:"pr"`
}

// ReassignRequest — запрос на переназначение ревьюера.
type ReassignRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

// ReassignResponse — ответ API после переназначения ревьюера.
type ReassignResponse struct {
	PR         PullRequestDTO `json:"pr"`
	ReplacedBy string         `json:"replaced_by"`
}

// PullRequestShortDTO — краткая информация о PR для /users/getReview.
type PullRequestShortDTO struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

// UserReviewResponse — ответ API со списком PR для ревью пользователя.
type UserReviewResponse struct {
	UserID       string                `json:"user_id"`
	PullRequests []PullRequestShortDTO `json:"pull_requests"`
}

// UserAssignmentStatDTO — статистика назначений на ревью по пользователю.
type UserAssignmentStatDTO struct {
	UserID      string `json:"user_id"`
	Assignments int64  `json:"assignments"`
}

// StatsAssignmentsResponse — ответ API со статистикой назначений.
type StatsAssignmentsResponse struct {
	Stats []UserAssignmentStatDTO `json:"stats"`
}
