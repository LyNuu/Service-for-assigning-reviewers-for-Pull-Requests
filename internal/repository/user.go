package repository

import (
	"avitoMerchStore/internal/model"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool}
}

func (r *UserRepository) GetById(ctx context.Context, id string) (*model.User, error) {
	u := &model.User{}
	err := r.pool.QueryRow(ctx, `
 		SELECT user_id, username, team_name, is_active 
        FROM reviewers
        WHERE user_id = $1
	`, id).Scan(
		&u.ID,
		&u.Name,
		&u.TeamName,
		&u.Status,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("%w", err)
	}
	return u, nil
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE reviewers 
        SET username = $1, team_name = $2, is_active = $3 
        WHERE user_id = $4
    `, user.Name, user.TeamName, user.Status, user.ID)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) GetPrById(ctx context.Context, id string) (*[]model.PullRequest, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
        SELECT EXISTS(SELECT 1 FROM reviewers WHERE user_id = $1)
    `, id).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	rows, err := r.pool.Query(ctx, `
        SELECT 
            pr.pull_request_id, 
            pr.pull_request_name, 
            pr.author_id, 
            pr.status
        FROM pull_requests pr
        INNER JOIN pull_request_reviewers prr ON pr.pull_request_id = prr.pull_request_id
        WHERE prr.user_id = $1
        ORDER BY pr.pull_request_id
    `, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pullRequests []model.PullRequest
	for rows.Next() {
		var pr model.PullRequest
		err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status)
		if err != nil {
			return nil, err
		}
		pullRequests = append(pullRequests, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &pullRequests, nil
}
