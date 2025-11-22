package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"pr-reviewer-service/internal/domain"
)

// PullRequestRepository реализует domain.PullRequestRepository для PostgreSQL.
type PullRequestRepository struct {
	db *sql.DB
}

// NewPullRequestRepository создаёт новый PullRequestRepository.
func NewPullRequestRepository(db *sql.DB) *PullRequestRepository {
	return &PullRequestRepository{db: db}
}

// Create создаёт pull request и его ревьюеров в одной транзакции.
func (r *PullRequestRepository) Create(ctx context.Context, pr domain.PullRequest) error {
	tx, err := r.db.BeginTx(ctx, nil)

	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO pull_requests (id, name, author_id, status, created_at, merged_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		pr.ID, pr.Name, pr.AuthorID, string(pr.Status), pr.CreatedAt, pr.MergedAt,
	)

	if err != nil {
		return fmt.Errorf("insert pull_request: %w", err)
	}

	for _, reviewerID := range pr.AssignedReviewers {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO pr_reviewers (pr_id, reviewer_id)
			 VALUES ($1, $2)`,
			pr.ID, reviewerID,
		); err != nil {
			return fmt.Errorf("insert pr_reviewer: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

// GetByID возвращает полный pull request с назначенными ревьюерами.
func (r *PullRequestRepository) GetByID(ctx context.Context, id string) (domain.PullRequest, error) {
	var pr domain.PullRequest

	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, author_id, status, created_at, merged_at
		   FROM pull_requests
		  WHERE id = $1`,
		id,
	).Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)

	if err == sql.ErrNoRows {
		return domain.PullRequest{}, domain.ErrNotFound
	}

	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("select pull_request: %w", err)
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT reviewer_id FROM pr_reviewers WHERE pr_id = $1`,
		id,
	)

	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("select reviewers: %w", err)
	}

	defer func() {
		_ = rows.Close()
	}()

	var reviewers []string

	for rows.Next() {
		var rid string

		if err := rows.Scan(&rid); err != nil {
			return domain.PullRequest{}, fmt.Errorf("scan reviewer: %w", err)
		}

		reviewers = append(reviewers, rid)
	}

	pr.AssignedReviewers = reviewers
	return pr, nil
}

// MarkMerged помечает pull request как merged и возвращает обновлённую сущность.
func (r *PullRequestRepository) MarkMerged(ctx context.Context, id string, mergedAt time.Time) (domain.PullRequest, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE pull_requests
		    SET status = $2,
		        merged_at = $3
		  WHERE id = $1`,
		id, string(domain.PRStatusMerged), mergedAt,
	)

	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("update pull_request merged: %w", err)
	}

	affected, err := res.RowsAffected()

	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("rows affected: %w", err)
	}

	if affected == 0 {
		return domain.PullRequest{}, domain.ErrNotFound
	}

	return r.GetByID(ctx, id)
}

// ReassignReviewer заменяет одного ревьюера другим в рамках транзакции.
func (r *PullRequestRepository) ReassignReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) (domain.PullRequest, error) {
	tx, err := r.db.BeginTx(ctx, nil)

	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("begin tx: %w", err)
	}

	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx,
		`DELETE FROM pr_reviewers WHERE pr_id = $1 AND reviewer_id = $2`,
		prID, oldReviewerID,
	)

	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("delete old reviewer: %w", err)
	}

	affected, err := res.RowsAffected()

	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("rows affected: %w", err)
	}

	if affected == 0 {
		return domain.PullRequest{}, domain.ErrReviewerNotAssigned
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO pr_reviewers (pr_id, reviewer_id)
		 VALUES ($1, $2)`,
		prID, newReviewerID,
	)

	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("insert new reviewer: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return domain.PullRequest{}, fmt.Errorf("commit tx: %w", err)
	}

	return r.GetByID(ctx, prID)
}

// ListByReviewer возвращает список PR, назначенных конкретному ревьюеру.
func (r *PullRequestRepository) ListByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequestShort, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT p.id, p.name, p.author_id, p.status
		   FROM pull_requests p
		   JOIN pr_reviewers rview ON p.id = rview.pr_id
		  WHERE rview.reviewer_id = $1`,
		reviewerID,
	)

	if err != nil {
		return nil, fmt.Errorf("select prs by reviewer: %w", err)
	}

	defer func() {
		_ = rows.Close()
	}()

	var res []domain.PullRequestShort

	for rows.Next() {
		var pr domain.PullRequestShort

		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status); err != nil {
			return nil, fmt.Errorf("scan pr: %w", err)
		}

		res = append(res, pr)
	}

	return res, nil
}

// PRExists проверяет, существует ли pull request с таким идентификатором.
func (r *PullRequestRepository) PRExists(ctx context.Context, id string) (bool, error) {
	var exists bool

	err := r.db.QueryRowContext(ctx,
		`SELECT TRUE FROM pull_requests WHERE id = $1`,
		id,
	).Scan(&exists)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("check pr exists: %w", err)
	}

	return true, nil
}

// GetAssignmentStatsByUser возвращает статистику количества назначений по каждому ревьюеру.
func (r *PullRequestRepository) GetAssignmentStatsByUser(ctx context.Context) ([]domain.AssignmentStatByUser, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT reviewer_id, COUNT(*) 
		   FROM pr_reviewers
		  GROUP BY reviewer_id`,
	)

	if err != nil {
		return nil, fmt.Errorf("select assignment stats: %w", err)
	}

	defer func() {
		_ = rows.Close()
	}()

	var res []domain.AssignmentStatByUser

	for rows.Next() {
		var s domain.AssignmentStatByUser

		if err := rows.Scan(&s.UserID, &s.Count); err != nil {
			return nil, fmt.Errorf("scan stat: %w", err)
		}

		res = append(res, s)
	}

	return res, nil
}

type txKey struct{}

// WithTx выполняет переданную функцию как транзакцию.
func (r *PullRequestRepository) WithTx(
	ctx context.Context,
	fn func(ctx context.Context, tx *sql.Tx) error,
) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)

	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}

		if err != nil {
			_ = tx.Rollback()

		} else {
			err = tx.Commit()
		}
	}()

	// Прокинем tx в контекст — если потом захочешь доставать его где-то ещё по ctx
	ctxWithTx := context.WithValue(ctx, txKey{}, tx)

	err = fn(ctxWithTx, tx)
	return err
}
