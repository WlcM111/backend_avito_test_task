package logging

import (
	"log/slog"
	"os"
)

// Logger — обёртка над slog.Logger.
type Logger struct {
	*slog.Logger
}

// NewLogger создаёт логгер в текстовом или JSON-формате в зависимости от окружения.
func NewLogger(env string) *Logger {
	var handler slog.Handler

	if env == "prod" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})

	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	return &Logger{slog.New(handler)}
}

// With возвращает логгер с добавленными полями контекста.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{l.Logger.With(args...)}
}
