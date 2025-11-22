package httpapi

import (
	"encoding/json"
	"net/http"

	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/internal/service"
)

// UserHandlers содержит HTTP-обработчики, связанные с пользователями.
type UserHandlers struct {
	svc *service.UserService
}

// NewUserHandlers создаёт набор HTTP-обработчиков для работы с пользователями.
func NewUserHandlers(svc *service.UserService) *UserHandlers {
	return &UserHandlers{svc: svc}
}

// SetIsActive обрабатывает запрос на смену флага активности пользователя.
func (h *UserHandlers) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var req SetIsActiveRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err)
		return
	}

	user, err := h.svc.SetIsActive(r.Context(), req.UserID, req.IsActive)

	if err != nil {
		WriteError(w, err)
		return
	}

	resp := SetIsActiveResponse{
		User: UserDTO{
			UserID:   user.ID,
			Username: user.Username,
			TeamName: user.TeamName,
			IsActive: user.IsActive,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// GetReviewPRs возвращает список pull request-ов, которые пользователь должен ревьюить.
func (h *UserHandlers) GetReviewPRs(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")

	if userID == "" {
		WriteError(w, &domain.DomainError{
			Code: domain.ErrorCodeNotFound,
			Err:  domain.ErrNotFound,
		})

		return
	}

	uid, prs, err := h.svc.GetReviewPRs(r.Context(), userID)

	if err != nil {
		WriteError(w, err)
		return
	}

	resp := UserReviewResponse{
		UserID:       uid,
		PullRequests: mapPRShortsToDTO(prs),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func mapPRShortsToDTO(prs []domain.PullRequestShort) []PullRequestShortDTO {
	res := make([]PullRequestShortDTO, 0, len(prs))

	for _, pr := range prs {
		res = append(res, PullRequestShortDTO{
			PullRequestID:   pr.ID,
			PullRequestName: pr.Name,
			AuthorID:        pr.AuthorID,
			Status:          string(pr.Status),
		})
	}

	return res
}
