package httpapi

import (
	"encoding/json"
	"net/http"

	"pr-reviewer-service/internal/service"
)

// StatsHandlers содержит HTTP-обработчики для статистики по назначенным ревью.
type StatsHandlers struct {
	svc *service.StatsService
}

// NewStatsHandlers создаёт набор HTTP-обработчиков статистики.
func NewStatsHandlers(svc *service.StatsService) *StatsHandlers {
	return &StatsHandlers{svc: svc}
}

// GetAssignmentsByUser возвращает статистику назначений на ревью по пользователям.
func (h *StatsHandlers) GetAssignmentsByUser(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.GetAssignmentsByUser(r.Context())

	if err != nil {
		WriteError(w, err)
		return
	}

	resp := StatsAssignmentsResponse{
		Stats: make([]UserAssignmentStatDTO, 0, len(stats)),
	}

	for _, s := range stats {
		resp.Stats = append(resp.Stats, UserAssignmentStatDTO{
			UserID:      s.UserID,
			Assignments: s.Count,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
