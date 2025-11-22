package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"pr-reviewer-service/internal/domain"
)

// TeamRepository реализует domain.TeamRepository для PostgreSQL.
type TeamRepository struct {
	db *sql.DB
}

// NewTeamRepository создаёт TeamRepository.
func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

// CreateTeam создаёт запись о команде.
func (r *TeamRepository) CreateTeam(ctx context.Context, name string, members []domain.User) error {
	now := time.Now().UTC()

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO teams (team_name, created_at, updated_at)
		 VALUES ($1, $2, $3)`,
		name, now, now,
	)

	if err != nil {
		return fmt.Errorf("insert team: %w", err)
	}

	return nil
}

// GetTeamWithMembers возвращает команду
func (r *TeamRepository) GetTeamWithMembers(ctx context.Context, teamName string) (domain.Team, error) {
	var t domain.Team

	err := r.db.QueryRowContext(ctx,
		`SELECT team_name FROM teams WHERE team_name = $1`,
		teamName,
	).Scan(&t.Name)

	if err == sql.ErrNoRows {
		return domain.Team{}, domain.ErrNotFound
	}

	if err != nil {
		return domain.Team{}, fmt.Errorf("select team: %w", err)
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT user_id, username, team_name, is_active, created_at, updated_at
		   FROM users
		  WHERE team_name = $1`,
		teamName,
	)

	if err != nil {
		return domain.Team{}, fmt.Errorf("select team members: %w", err)
	}

	defer func() {
		_ = rows.Close()
	}()

	var members []domain.User

	for rows.Next() {
		var u domain.User

		if err := rows.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return domain.Team{}, fmt.Errorf("scan team member: %w", err)
		}

		members = append(members, u)
	}

	t.Members = members
	return t, nil
}

// TeamExists проверяет что команда создана
func (r *TeamRepository) TeamExists(ctx context.Context, name string) (bool, error) {
	var exists bool

	err := r.db.QueryRowContext(ctx,
		`SELECT TRUE FROM teams WHERE team_name = $1`,
		name,
	).Scan(&exists)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("check team exists: %w", err)
	}

	return true, nil
}
