package httpapi

import (
	"encoding/json"
	"net/http"

	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/internal/service"
)

// TeamHandlers содержит HTTP-обработчики, связанные с командами.
type TeamHandlers struct {
	svc *service.TeamService
}

// NewTeamHandlers создаёт новый набор обработчиков команд.
func NewTeamHandlers(svc *service.TeamService) *TeamHandlers {
	return &TeamHandlers{svc: svc}
}

// CreateTeam обрабатывает создание команды.
func (h *TeamHandlers) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req TeamRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err)
		return
	}

	members := make([]domain.User, 0, len(req.Members))

	for _, m := range req.Members {
		members = append(members, domain.User{
			ID:       m.UserID,
			Username: m.Username,
			TeamName: req.TeamName,
			IsActive: m.IsActive,
		})
	}

	team, err := h.svc.CreateTeam(r.Context(), req.TeamName, members)

	if err != nil {
		WriteError(w, err)
		return
	}

	resp := TeamCreateResponse{
		Team: TeamDTO{
			TeamName: team.Name,
			Members:  mapUsersToTeamMembers(team.Members),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

// GetTeam возвращает описание команды по имени.
func (h *TeamHandlers) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")

	if teamName == "" {
		WriteError(w, &domain.DomainError{
			Code: domain.ErrorCodeNotFound,
			Err:  domain.ErrNotFound,
		})

		return
	}

	team, err := h.svc.GetTeam(r.Context(), teamName)

	if err != nil {
		WriteError(w, err)
		return
	}

	resp := TeamDTO{
		TeamName: team.Name,
		Members:  mapUsersToTeamMembers(team.Members),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func mapUsersToTeamMembers(users []domain.User) []TeamMemberDTO {
	res := make([]TeamMemberDTO, 0, len(users))

	for _, u := range users {
		res = append(res, TeamMemberDTO{
			UserID:   u.ID,
			Username: u.Username,
			IsActive: u.IsActive,
		})
	}

	return res
}
