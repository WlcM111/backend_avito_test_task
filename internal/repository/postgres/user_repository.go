package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"pr-reviewer-service/internal/domain"
)

// UserRepository реализует domain.UserRepository для PostgreSQL.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository создаёт новый UserRepository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByID возвращает пользователя по его идентификатору.
func (r *UserRepository) GetByID(ctx context.Context, id string) (domain.User, error) {
	var u domain.User

	err := r.db.QueryRowContext(ctx,
		`SELECT user_id, username, team_name, is_active, created_at, updated_at
		   FROM users WHERE user_id = $1`,
		id,
	).Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)

	if err == sql.ErrNoRows {
		return domain.User{}, domain.ErrNotFound
	}

	if err != nil {
		return domain.User{}, fmt.Errorf("select user: %w", err)
	}

	return u, nil
}

// UpsertUsers создаёт или обновляет пользователей в составе команды.
func (r *UserRepository) UpsertUsers(ctx context.Context, teamName string, users []domain.User) error {
	now := time.Now().UTC()

	for _, u := range users {
		if _, err := r.db.ExecContext(ctx,
			`INSERT INTO users (user_id, username, team_name, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 ON CONFLICT (user_id) DO UPDATE
			 SET username = EXCLUDED.username,
			     team_name = EXCLUDED.team_name,
			     is_active = EXCLUDED.is_active,
			     updated_at = EXCLUDED.updated_at`,
			u.ID, u.Username, teamName, u.IsActive, now, now,
		); err != nil {
			return fmt.Errorf("upsert user %s: %w", u.ID, err)
		}
	}

	return nil
}

// SetIsActive изменяет флаг активности пользователя и возвращает обновлённые данные.
func (r *UserRepository) SetIsActive(ctx context.Context, id string, isActive bool) (domain.User, error) {
	var u domain.User

	err := r.db.QueryRowContext(ctx,
		`UPDATE users
		    SET is_active = $2,
		        updated_at = $3
		  WHERE user_id = $1
	      RETURNING user_id, username, team_name, is_active, created_at, updated_at`,
		id, isActive, time.Now().UTC(),
	).Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)

	if err == sql.ErrNoRows {
		return domain.User{}, domain.ErrNotFound
	}

	if err != nil {
		return domain.User{}, fmt.Errorf("update user is_active: %w", err)
	}

	return u, nil
}

// GetActiveTeamMembersExcept возвращает активных участников команды кроме указанного пользователя.
func (r *UserRepository) GetActiveTeamMembersExcept(ctx context.Context, teamName, excludeUserID string) ([]domain.User, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT user_id, username, team_name, is_active, created_at, updated_at
		   FROM users
		  WHERE team_name = $1
		    AND is_active = TRUE
		    AND user_id <> $2`,
		teamName, excludeUserID,
	)

	if err != nil {
		return nil, fmt.Errorf("select active team members: %w", err)
	}

	defer func() {
		_ = rows.Close()
	}()

	var res []domain.User

	for rows.Next() {
		var u domain.User

		if err := rows.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}

		res = append(res, u)
	}

	return res, nil
}

// GetTeamByUserID возвращает имя команды по идентификатору пользователя.
func (r *UserRepository) GetTeamByUserID(ctx context.Context, userID string) (string, error) {
	var teamName string

	err := r.db.QueryRowContext(ctx,
		`SELECT team_name FROM users WHERE user_id = $1`,
		userID,
	).Scan(&teamName)

	if err == sql.ErrNoRows {
		return "", domain.ErrNotFound
	}

	if err != nil {
		return "", fmt.Errorf("select team by user: %w", err)
	}

	return teamName, nil
}
