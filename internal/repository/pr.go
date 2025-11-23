package repository

import (
	"avitoMerchStore/internal/model"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PrRepository struct {
	pool *pgxpool.Pool
}

func NewPrRepository(pool *pgxpool.Pool) *PrRepository {
	return &PrRepository{pool: pool}
}

func (r *PrRepository) CreatePR(ctx context.Context, pr *model.PullRequest) (*model.PullRequest, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	err = r.CheckExists(ctx, pr.ID)
	if err != nil {
		if errors.Is(err, ErrPullRequestAlreadyExists) {
			return nil, ErrPullRequestAlreadyExists
		}
		return nil, err
	}
	rew, err := r.getReviewers(ctx, tx, pr.AuthorID)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx, `
        INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status)
        VALUES ($1, $2, $3, $4)
    `, pr.ID, pr.Name, pr.AuthorID, pr.Status)
	if err != nil {
		return nil, err
	}

	for _, reviewerID := range *rew {
		_, err = tx.Exec(ctx, `
            INSERT INTO pull_request_reviewers (pull_request_id, user_id)
            VALUES ($1, $2)
        `, pr.ID, reviewerID)
		if err != nil {
			return nil, err
		}
	}
	pr.Reviewers = *rew
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return pr, nil
}

func (r *PrRepository) CheckExists(ctx context.Context, id string) error {
	var flg bool
	err := r.pool.QueryRow(ctx, `
        SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)
    `, id).Scan(&flg)
	if err != nil {
		return err
	}
	if flg {
		return ErrPullRequestAlreadyExists
	}
	return nil
}

func (r *PrRepository) getReviewers(ctx context.Context, tx pgx.Tx, authorId string) (*[]string, error) {
	var authorTeam string
	var isActive bool
	err := tx.QueryRow(ctx, `
        SELECT team_name, is_active 
        FROM reviewers 
        WHERE user_id = $1
    `, authorId).Scan(&authorTeam, &isActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAuthorNotFound
		}
		return nil, err
	}
	if !isActive {
		return nil, ErrAuthorIsNotActive
	}

	rows, err := tx.Query(ctx, `
        SELECT user_id 
        FROM reviewers 
        WHERE team_name = $1 
          AND user_id != $2 
          AND is_active = true
        LIMIT 2
    `, authorTeam, authorId)
	if err != nil {
		return nil, ErrNotReviewersInTeam
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var reviewer string
		if err := rows.Scan(&reviewer); err != nil {
			return nil, ErrNotReviewersInTeam
		}
		reviewers = append(reviewers, reviewer)
	}
	if err := rows.Err(); err != nil {
		return nil, ErrNotReviewersInTeam
	}
	return &reviewers, nil
}

func (r *PrRepository) MergePR(ctx context.Context, prID string) (*model.PullRequest, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var pr model.PullRequest
	err = tx.QueryRow(ctx, `
        SELECT pull_request_id, pull_request_name, author_id, status
        FROM pull_requests WHERE pull_request_id = $1
    `, prID).Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPullRequestNotFound
		}
		return nil, err
	}

	if pr.Status == model.StatusMerged {
		reviewers, err := r.getReviewers(ctx, tx, pr.AuthorID)
		if err != nil {
			return nil, err
		}
		pr.Reviewers = *reviewers
		return &pr, nil
	}

	_, err = tx.Exec(ctx, `
        UPDATE pull_requests 
        SET status = $1
        WHERE pull_request_id = $2
    `, model.StatusMerged, prID)

	if err != nil {
		return nil, err
	}

	reviewers, err := r.getReviewers(ctx, tx, pr.AuthorID)
	if err != nil {
		return nil, err
	}
	pr.Status = model.StatusMerged
	pr.Reviewers = *reviewers

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &pr, nil
}

func (r *PrRepository) ReassignPR(ctx context.Context, pr *model.PullRequest, oldId string) (*model.PullRequest, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var p model.PullRequest
	err = tx.QueryRow(ctx, `
        SELECT pull_request_id, pull_request_name, author_id, status
        FROM pull_requests WHERE pull_request_id = $1
    `, pr.ID).Scan(&p.ID, &p.Name, &p.AuthorID, &p.Status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPullRequestNotFound
		}
		return nil, err
	}
	if p.Status == model.StatusMerged {
		return nil, ErrPullRequestAlreadyMerged
	}
	var isAssigned bool
	err = tx.QueryRow(ctx, `
        SELECT EXISTS(
            SELECT 1 FROM pull_request_reviewers 
            WHERE pull_request_id = $1 AND user_id = $2
        )
    `, pr.ID, oldId).Scan(&isAssigned)
	if err != nil {
		return nil, err
	}
	if !isAssigned {
		return nil, ErrReviewerIsNotAssigned
	}

	var oldUserTeam string
	err = tx.QueryRow(ctx, `
        SELECT team_name 
        FROM reviewers 
        WHERE user_id = $1 AND is_active = true
    `, oldId).Scan(&oldUserTeam)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAuthorNotFound
		}
		return nil, err
	}

	var newReviewerID string
	err = tx.QueryRow(ctx, `
        SELECT user_id 
        FROM reviewers 
        WHERE team_name = $1 
          AND user_id != $2 
          AND user_id != $3 
          AND is_active = true
        LIMIT 1
    `, oldUserTeam, p.AuthorID, oldId).Scan(&newReviewerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTeamNotFound
		}
		return nil, err
	}

	_, err = tx.Exec(ctx, `
        UPDATE pull_request_reviewers 
        SET user_id = $1
        WHERE pull_request_id = $2 AND user_id = $3
    `, newReviewerID, pr.ID, oldId)
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, `
        SELECT user_id 
        FROM pull_request_reviewers 
        WHERE pull_request_id = $1
    `, pr.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var reviewer string
		if err := rows.Scan(&reviewer); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, reviewer)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	p.Reviewers = reviewers
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &p, nil
}
