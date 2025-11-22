package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pr-reviewer-service/internal/config"
	httpapi "pr-reviewer-service/internal/http"
	"pr-reviewer-service/internal/logging"
	"pr-reviewer-service/internal/random"
	"pr-reviewer-service/internal/repository/postgres"
	"pr-reviewer-service/internal/server"
	"pr-reviewer-service/internal/service"
	"pr-reviewer-service/internal/storage"
)

func main() {
	// Load config
	cfg, err := config.Load()

	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Init logger
	logger := logging.NewLogger(cfg.Env)
	logger.Info("starting service", "env", cfg.Env)

	// Init DB
	db, err := postgres.NewDB(cfg.DB)

	if err != nil {
		logger.Error("failed to connect to db", "err", err)
		os.Exit(1)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close db", "err", err)
		}
	}()

	// Run migrations
	if err := storage.RunMigrations(db, "migrations"); err != nil {
		logger.Error("failed to run migrations", "err", err)
		os.Exit(1)
	}

	// Repositories
	teamRepo := postgres.NewTeamRepository(db)
	userRepo := postgres.NewUserRepository(db)
	prRepo := postgres.NewPullRequestRepository(db)

	// Random source
	randSource := random.NewCryptoRand()

	// Services
	teamSvc := service.NewTeamService(teamRepo, userRepo)
	userSvc := service.NewUserService(userRepo, prRepo)
	prSvc := service.NewPullRequestService(prRepo, userRepo, teamRepo, randSource)
	statsSvc := service.NewStatsService(prRepo)

	// HTTP router
	router := httpapi.NewRouter(teamSvc, userSvc, prSvc, statsSvc, logger)

	// HTTP server
	httpServer := server.NewHTTPServer(cfg.HTTP, router, logger)

	// Graceful shutdown
	go func() {
		if err := httpServer.Start(); err != nil {
			logger.Error("http server error", "err", err)
		}
	}()

	logger.Info("server started", "addr", cfg.HTTP.Addr())

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	logger.Info("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)

	} else {
		logger.Info("server stopped gracefully")
	}
}
