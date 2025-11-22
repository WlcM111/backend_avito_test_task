package httpapi

import (
	"net/http"
	"time"

	"pr-reviewer-service/internal/logging"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware логирует входящие HTTP-запросы и их статус/длительность.
func LoggingMiddleware(logger *logging.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)
			duration := time.Since(start)

			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"duration_ms", duration.Milliseconds(),
			)
		})
	}
}

// RecoveryMiddleware перехватывает panic, логирует их и возвращает INTERNAL ошибку.
func RecoveryMiddleware(logger *logging.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic recovered", "panic", rec)
					WriteError(w, &domainErrorInternal{})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// domainErrorInternal используется для возврата INTERNAL при панике
type domainErrorInternal struct{}

func (d *domainErrorInternal) Error() string { return "internal error" }
