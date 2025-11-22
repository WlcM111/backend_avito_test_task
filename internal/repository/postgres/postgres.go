package postgres

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"pr-reviewer-service/internal/config"
)

// NewDB создаёт и настраивает подключение к PostgreSQL.
func NewDB(cfg config.DBConfig) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DSN)

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}
