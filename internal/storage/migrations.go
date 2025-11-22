package storage

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// RunMigrations применяет SQL-миграции из указанной директории.
func RunMigrations(db *sql.DB, dir string) error {
	if err := ensureMigrationsTable(db); err != nil {
		return fmt.Errorf("ensure migrations table: %w", err)
	}

	entries, err := os.ReadDir(dir)

	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		if filepath.Ext(e.Name()) == ".sql" {
			files = append(files, e.Name())
		}
	}

	sort.Strings(files)

	for _, name := range files {
		applied, err := isMigrationApplied(db, name)

		if err != nil {
			return err
		}

		if applied {
			continue
		}

		path := filepath.Join(dir, name)
		content, err := ioutil.ReadFile(path)

		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		tx, err := db.Begin()

		if err != nil {
			return fmt.Errorf("begin tx for migration %s: %w", name, err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("exec migration %s: %w", name, err)
		}

		if _, err := tx.Exec(
			`INSERT INTO schema_migrations (name, applied_at) VALUES ($1, $2)`,
			name,
			time.Now().UTC(),
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", name, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
	}

	return nil
}

func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS schema_migrations (
  id         SERIAL PRIMARY KEY,
  name       TEXT NOT NULL UNIQUE,
  applied_at TIMESTAMP WITH TIME ZONE NOT NULL
)`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	return nil
}

func isMigrationApplied(db *sql.DB, name string) (bool, error) {
	var exists bool
	err := db.QueryRow(`SELECT TRUE FROM schema_migrations WHERE name = $1`, name).Scan(&exists)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("check migration %s: %w", name, err)
	}

	return true, nil
}
