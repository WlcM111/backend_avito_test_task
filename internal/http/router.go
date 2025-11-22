package httpapi

import (
	nethttp "net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"pr-reviewer-service/internal/logging"
	"pr-reviewer-service/internal/service"
)

// NewRouter настраивает HTTP-маршруты и middleware сервиса.
func NewRouter(
	teamSvc *service.TeamService,
	userSvc *service.UserService,
	prSvc *service.PullRequestService,
	statsSvc *service.StatsService,
	logger *logging.Logger,
) nethttp.Handler {
	r := chi.NewRouter()

	r.Use(LoggingMiddleware(logger))
	r.Use(RecoveryMiddleware(logger))

	teamHandlers := NewTeamHandlers(teamSvc)
	userHandlers := NewUserHandlers(userSvc)
	prHandlers := NewPullRequestHandlers(prSvc)
	statsHandlers := NewStatsHandlers(statsSvc)

	r.Get("/health", HealthHandler)

	r.Route("/team", func(r chi.Router) {
		r.Post("/add", teamHandlers.CreateTeam)
		r.Get("/get", teamHandlers.GetTeam)
	})

	r.Route("/users", func(r chi.Router) {
		r.Post("/setIsActive", userHandlers.SetIsActive)
		r.Get("/getReview", userHandlers.GetReviewPRs)
	})

	r.Route("/pullRequest", func(r chi.Router) {
		r.Post("/create", prHandlers.CreatePR)
		r.Post("/merge", prHandlers.MergePR)
		r.Post("/reassign", prHandlers.ReassignReviewer)
	})

	// Доп. статистика
	r.Get("/stats/assignments", statsHandlers.GetAssignmentsByUser)

	// Оборачиваем в TimeoutHandler, чтобы приблизиться к SLI 300ms
	timeout := 250 * time.Millisecond
	return nethttp.TimeoutHandler(r, timeout, `{"error":{"code":"INTERNAL","message":"request timeout"}}`)
}
