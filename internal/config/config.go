package config

import (
	"fmt"
	"os"
	"time"
)

// HTTPConfig описывает настройки HTTP-сервера.
type HTTPConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// Addr возвращает адрес для http.Server (с двоеточием перед портом).
func (h HTTPConfig) Addr() string {
	if h.Port == "" {
		return ":8080"
	}

	// Разрешить порты ":8080" and "8080"
	if h.Port[0] == ':' {
		return h.Port
	}

	return fmt.Sprintf(":%s", h.Port)
}

// DBConfig хранит настройки доступа к базе данных.
type DBConfig struct {
	DSN string
}

// Config объединяет все настройки сервиса.
type Config struct {
	HTTP HTTPConfig
	DB   DBConfig
	Env  string
}

// Load загружает конфигурацию из переменных окружения.
func Load() (*Config, error) {
	httpPort := getenv("HTTP_PORT", "8080")
	dbDSN := os.Getenv("DB_DSN")

	if dbDSN == "" {
		dbDSN = "postgres://postgres:postgres@postgres:5432/postgres?sslmode=disable"
	}

	env := getenv("ENV", "dev")

	return &Config{
		HTTP: HTTPConfig{
			Port:         httpPort,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		DB: DBConfig{
			DSN: dbDSN,
		},
		Env: env,
	}, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return def
}
