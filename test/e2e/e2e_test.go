package e2e

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"pr-reviewer-service/internal/config"
	httpapi "pr-reviewer-service/internal/http"
	"pr-reviewer-service/internal/logging"
	"pr-reviewer-service/internal/random"
	"pr-reviewer-service/internal/repository/postgres"
	"pr-reviewer-service/internal/service"
	"pr-reviewer-service/internal/storage"
)

type teamCreateResp struct {
	Team struct {
		TeamName string `json:"team_name"`
		Members  []struct {
			UserID   string `json:"user_id"`
			Username string `json:"username"`
			IsActive bool   `json:"is_active"`
		} `json:"members"`
	} `json:"team"`
}

type createPRResp struct {
	PR pullRequestDTO `json:"pr"`
}

type mergePRResp struct {
	PR pullRequestDTO `json:"pr"`
}

type reassignResp struct {
	PR         pullRequestDTO `json:"pr"`
	ReplacedBy string         `json:"replaced_by"`
}

type pullRequestDTO struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         *time.Time `json:"createdAt"`
	MergedAt          *time.Time `json:"mergedAt"`
}

type errorResp struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type userReviewResp struct {
	UserID       string `json:"user_id"`
	PullRequests []struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
		Status          string `json:"status"`
	} `json:"pull_requests"`
}

type statsResp struct {
	Stats []struct {
		UserID      string `json:"user_id"`
		Assignments int64  `json:"assignments"`
	} `json:"stats"`
}

type testEnv struct {
	t      *testing.T
	db     *sql.DB
	server *httptest.Server
	client *http.Client
	base   string
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	// DSN для тестовой БД
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	dbCfg := config.DBConfig{DSN: dsn}
	db, err := postgres.NewDB(dbCfg)

	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	// Миграции
	if err := storage.RunMigrations(db, "migrations"); err != nil {
		_ = db.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	cleanDB(t, db)

	teamRepo := postgres.NewTeamRepository(db)
	userRepo := postgres.NewUserRepository(db)
	prRepo := postgres.NewPullRequestRepository(db)

	randSource := random.NewCryptoRand()
	logger := logging.NewLogger("test")

	teamSvc := service.NewTeamService(teamRepo, userRepo)
	userSvc := service.NewUserService(userRepo, prRepo)
	prSvc := service.NewPullRequestService(prRepo, userRepo, teamRepo, randSource)
	statsSvc := service.NewStatsService(prRepo)

	router := httpapi.NewRouter(teamSvc, userSvc, prSvc, statsSvc, logger)
	ts := httptest.NewServer(router)

	return &testEnv{
		t:      t,
		db:     db,
		server: ts,
		client: ts.Client(),
		base:   ts.URL,
	}
}

func (env *testEnv) teardown() {
	_ = env.db.Close()
	env.server.Close()
}

func cleanDB(t *testing.T, db *sql.DB) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tables := []string{"pr_reviewers", "pull_requests", "users", "teams"}

	for _, tbl := range tables {
		if _, err := db.ExecContext(ctx, "DELETE FROM "+tbl); err != nil {
			t.Fatalf("failed to clean table %s: %v", tbl, err)
		}
	}
}

// ==== Хелперы HTTP-запросов ====

func (env *testEnv) postJSON(path string, reqBody any, expectedStatus int, out any) {
	env.t.Helper()

	var bodyBytes []byte

	if reqBody != nil {
		var err error
		bodyBytes, err = json.Marshal(reqBody)

		if err != nil {
			env.t.Fatalf("failed to marshal request: %v", err)
		}
	}
	req, err := http.NewRequest(http.MethodPost, env.base+path, bytes.NewReader(bodyBytes))

	if err != nil {
		env.t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := env.client.Do(req)

	if err != nil {
		env.t.Fatalf("request failed: %v", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != expectedStatus {
		var errBody errorResp
		_ = json.NewDecoder(resp.Body).Decode(&errBody)

		env.t.Fatalf("unexpected status for POST %s: got %d, want %d, error=%+v",
			path, resp.StatusCode, expectedStatus, errBody)
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			env.t.Fatalf("failed to decode response for %s: %v", path, err)
		}
	}
}

func (env *testEnv) get(path string, expectedStatus int, out any) {
	env.t.Helper()

	req, err := http.NewRequest(http.MethodGet, env.base+path, nil)

	if err != nil {
		env.t.Fatalf("failed to create GET request: %v", err)
	}

	resp, err := env.client.Do(req)

	if err != nil {
		env.t.Fatalf("GET request failed: %v", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != expectedStatus {
		var errBody errorResp
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		env.t.Fatalf("unexpected status for GET %s: got %d, want %d, error=%+v",
			path, resp.StatusCode, expectedStatus, errBody)
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			env.t.Fatalf("failed to decode response for %s: %v", path, err)
		}
	}
}

// ==== E2E-тесты ====

func TestEndToEnd_HappyFlow(t *testing.T) {
	env := setupTestEnv(t)
	defer env.teardown()

	// 1. health
	var health struct {
		Status string `json:"status"`
	}

	env.get("/health", http.StatusOK, &health)

	if health.Status != "ok" {
		t.Fatalf("unexpected health status: %s", health.Status)
	}

	// 2. создаём команду backend с тремя активными участниками
	teamReq := map[string]any{
		"team_name": "backend",
		"members": []map[string]any{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Eve", "is_active": true},
		},
	}

	var teamResp teamCreateResp
	env.postJSON("/team/add", teamReq, http.StatusCreated, &teamResp)

	if teamResp.Team.TeamName != "backend" {
		t.Fatalf("unexpected team_name: %s", teamResp.Team.TeamName)
	}

	if len(teamResp.Team.Members) != 3 {
		t.Fatalf("expected 3 team members, got %d", len(teamResp.Team.Members))
	}

	// 3. создаём PR от u1
	prID := "pr-e2e-1"
	createReq := map[string]any{
		"pull_request_id":   prID,
		"pull_request_name": "Add search",
		"author_id":         "u1",
	}

	var prCreate createPRResp
	env.postJSON("/pullRequest/create", createReq, http.StatusCreated, &prCreate)
	pr := prCreate.PR

	if pr.PullRequestID != prID {
		t.Fatalf("unexpected pr id: %s", pr.PullRequestID)
	}

	if pr.Status != "OPEN" {
		t.Fatalf("expected status OPEN, got %s", pr.Status)
	}

	if len(pr.AssignedReviewers) == 0 || len(pr.AssignedReviewers) > 2 {
		t.Fatalf("expected 1 or 2 assigned reviewers, got %d", len(pr.AssignedReviewers))
	}

	for _, rid := range pr.AssignedReviewers {
		if rid == "u1" {
			t.Fatalf("author must not be assigned as reviewer")
		}
	}

	// 4. переназначаем одного из ревьюверов
	oldReviewer := pr.AssignedReviewers[0]
	reassignReq := map[string]any{
		"pull_request_id": prID,
		"old_user_id":     oldReviewer,
	}

	var reassign reassignResp
	env.postJSON("/pullRequest/reassign", reassignReq, http.StatusOK, &reassign)

	if reassign.PR.Status != "OPEN" {
		t.Fatalf("expected status OPEN after reassign, got %s", reassign.PR.Status)
	}

	if len(reassign.PR.AssignedReviewers) != 2 {
		t.Fatalf("expected 2 reviewers after reassign, got %d", len(reassign.PR.AssignedReviewers))
	}

	if reassign.ReplacedBy == oldReviewer {
		t.Fatalf("replaced_by must differ from old reviewer")
	}

	// убедимся, что старый ревьювер больше не в списке
	for _, rid := range reassign.PR.AssignedReviewers {
		if rid == oldReviewer {
			t.Fatalf("old reviewer still assigned after reassign")
		}
	}

	// 5. merge PR (идемпотентный)
	mergeReq := map[string]any{
		"pull_request_id": prID,
	}

	var mergeResp mergePRResp
	env.postJSON("/pullRequest/merge", mergeReq, http.StatusOK, &mergeResp)

	if mergeResp.PR.Status != "MERGED" {
		t.Fatalf("expected status MERGED, got %s", mergeResp.PR.Status)
	}

	if mergeResp.PR.MergedAt == nil {
		t.Fatalf("expected mergedAt to be set")
	}

	// повторный merge — должен вернуть тот же MERGED без ошибки
	var mergeResp2 mergePRResp
	env.postJSON("/pullRequest/merge", mergeReq, http.StatusOK, &mergeResp2)

	if mergeResp2.PR.Status != "MERGED" {
		t.Fatalf("expected status MERGED on second merge, got %s", mergeResp2.PR.Status)
	}

	// 6. попытка reassign после MERGED → PR_MERGED
	var errBody errorResp
	env.postJSON("/pullRequest/reassign", reassignReq, http.StatusConflict, &errBody)

	if errBody.Error.Code != "PR_MERGED" {
		t.Fatalf("expected error code PR_MERGED, got %s", errBody.Error.Code)
	}

	// 7. /users/getReview для одного из ревьюверов
	reviewerForCheck := reassign.PR.AssignedReviewers[0]
	var reviewResp userReviewResp
	env.get("/users/getReview?user_id="+reviewerForCheck, http.StatusOK, &reviewResp)

	if reviewResp.UserID == "" {
		t.Fatalf("empty user_id in getReview response")
	}

	found := false

	for _, p := range reviewResp.PullRequests {
		if p.PullRequestID == prID {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected PR %s in /users/getReview for %s", prID, reviewerForCheck)
	}

	// 8. /stats/assignments — статистика по ревьюверам
	var stats statsResp
	env.get("/stats/assignments", http.StatusOK, &stats)

	if len(stats.Stats) == 0 {
		t.Fatalf("expected non-empty stats")
	}

	hasSomeone := false

	for _, s := range stats.Stats {
		if s.Assignments > 0 {
			hasSomeone = true
			break
		}
	}

	if !hasSomeone {
		t.Fatalf("expected at least one user with assignments > 0")
	}
}

// Тест на кейс "мало кандидатов": только один активный ревьювер кроме автора.
func TestEndToEnd_SingleReviewerCandidate(t *testing.T) {
	env := setupTestEnv(t)
	defer env.teardown()

	teamReq := map[string]any{
		"team_name": "single",
		"members": []map[string]any{
			{"user_id": "a1", "username": "Author", "is_active": true},
			{"user_id": "r1", "username": "OnlyReviewer", "is_active": true},
			{"user_id": "i1", "username": "Inactive", "is_active": false},
		},
	}

	var teamResp teamCreateResp
	env.postJSON("/team/add", teamReq, http.StatusCreated, &teamResp)

	createReq := map[string]any{
		"pull_request_id":   "pr-single-1",
		"pull_request_name": "Single candidate",
		"author_id":         "a1",
	}

	var prCreate createPRResp
	env.postJSON("/pullRequest/create", createReq, http.StatusCreated, &prCreate)

	if len(prCreate.PR.AssignedReviewers) != 1 {
		t.Fatalf("expected exactly 1 reviewer (only one active candidate), got %d", len(prCreate.PR.AssignedReviewers))
	}

	if prCreate.PR.AssignedReviewers[0] != "r1" {
		t.Fatalf("expected reviewer r1, got %v", prCreate.PR.AssignedReviewers[0])
	}
}

// Дополнительный тест: нет кандидатов (команда = только автор → 0 ревьюверов).
func TestEndToEnd_NoReviewerCandidates(t *testing.T) {
	env := setupTestEnv(t)
	defer env.teardown()

	teamReq := map[string]any{
		"team_name": "no-reviewers",
		"members": []map[string]any{
			{"user_id": "a1", "username": "AuthorOnly", "is_active": true},
		},
	}

	var teamResp teamCreateResp
	env.postJSON("/team/add", teamReq, http.StatusCreated, &teamResp)

	createReq := map[string]any{
		"pull_request_id":   "pr-no-reviewer-1",
		"pull_request_name": "No reviewers",
		"author_id":         "a1",
	}

	var prCreate createPRResp
	env.postJSON("/pullRequest/create", createReq, http.StatusCreated, &prCreate)

	if len(prCreate.PR.AssignedReviewers) != 0 {
		t.Fatalf("expected 0 reviewers (no candidates), got %d", len(prCreate.PR.AssignedReviewers))
	}
}
