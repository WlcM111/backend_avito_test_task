package server

import (
	"context"
	"net/http"
	"time"

	"pr-reviewer-service/internal/config"
	"pr-reviewer-service/internal/logging"
)

// HTTPServer оборачивает http.Server с доп. логикой.
type HTTPServer struct {
	srv    *http.Server
	logger *logging.Logger
}

// NewHTTPServer создаёт HTTP-сервер с заданной конфигурацией.
func NewHTTPServer(cfg config.HTTPConfig, handler http.Handler, logger *logging.Logger) *HTTPServer {
	srv := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return &HTTPServer{
		srv:    srv,
		logger: logger,
	}
}

// Start запускает HTTP-сервер.
func (s *HTTPServer) Start() error {
	return s.srv.ListenAndServe()
}

// Shutdown выполняет graceful shutdown HTTP-сервера.
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	shutdownCh := make(chan error, 1)

	go func() {
		shutdownCh <- s.srv.Shutdown(ctx)
	}()

	select {
	case err := <-shutdownCh:
		return err

	case <-ctx.Done():
		return ctx.Err()

	case <-time.After(10 * time.Second):
		return context.DeadlineExceeded
	}
}
