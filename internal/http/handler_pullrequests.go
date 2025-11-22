package httpapi

import (
	"encoding/json"
	"net/http"

	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/internal/service"
)

// PullRequestHandlers содержит HTTP-обработчики, связанные с pull request-ами.
type PullRequestHandlers struct {
	svc *service.PullRequestService
}

// NewPullRequestHandlers создаёт набор HTTP-обработчиков для работы с pull request-ами.
func NewPullRequestHandlers(svc *service.PullRequestService) *PullRequestHandlers {
	return &PullRequestHandlers{svc: svc}
}

// CreatePR обрабатывает запрос на создание нового pull request.
func (h *PullRequestHandlers) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req CreatePRRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err)
		return
	}

	pr, err := h.svc.CreatePR(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)

	if err != nil {
		WriteError(w, err)
		return
	}

	resp := CreatePRResponse{
		PR: mapPRToDTO(pr),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

// MergePR обрабатывает запрос на пометку pull request как merged.
func (h *PullRequestHandlers) MergePR(w http.ResponseWriter, r *http.Request) {
	var req MergePRRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err)
		return
	}

	pr, err := h.svc.MergePR(r.Context(), req.PullRequestID)

	if err != nil {
		WriteError(w, err)
		return
	}

	resp := MergePRResponse{
		PR: mapPRToDTO(pr),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// ReassignReviewer обрабатывает запрос на переназначение ревьюера для pull request.
func (h *PullRequestHandlers) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	var req ReassignRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err)
		return
	}

	pr, replacedBy, err := h.svc.ReassignReviewer(r.Context(), req.PullRequestID, req.OldUserID)

	if err != nil {
		WriteError(w, err)
		return
	}

	resp := ReassignResponse{
		PR:         mapPRToDTO(pr),
		ReplacedBy: replacedBy,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func mapPRToDTO(pr domain.PullRequest) PullRequestDTO {
	return PullRequestDTO{
		PullRequestID:     pr.ID,
		PullRequestName:   pr.Name,
		AuthorID:          pr.AuthorID,
		Status:            string(pr.Status),
		AssignedReviewers: pr.AssignedReviewers,
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}
