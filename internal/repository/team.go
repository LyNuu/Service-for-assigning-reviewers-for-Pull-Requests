package repository

import (
	"avitoMerchStore/internal/model"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamRepository struct {
	pool *pgxpool.Pool
}

func NewTeamRepository(pool *pgxpool.Pool) *TeamRepository {
	return &TeamRepository{pool: pool}
}

func (r *TeamRepository) AddTeam(ctx context.Context, team *model.Team) (*model.Team, error) {
	err := r.CheckNotExistsTeam(ctx, team.Team)
	if err != nil {
		return nil, ErrTeamAlreadyExists
	}
	_, err = r.pool.Exec(ctx, `
        INSERT INTO teams (team_name) VALUES ($1)
    `, team.Team)
	if err != nil {
		return nil, err
	}
	for _, u := range team.Users {
		_, err = r.pool.Exec(ctx, `
			INSERT INTO reviewers (user_id, username, team_name, is_active) 
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id) DO UPDATE SET 
				username = EXCLUDED.username,
				team_name = EXCLUDED.team_name,
				is_active = EXCLUDED.is_active
		`, u.ID, u.Name, team.Team, u.Status)
	}
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (r *TeamRepository) CheckNotExistsTeam(ctx context.Context, teamName string) error {
	var flg bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)
		`, teamName).Scan(&flg)
	if err != nil {
		return err
	}
	if flg {
		return ErrTeamAlreadyExists
	}
	return nil
}

func (r *TeamRepository) GetTeamByName(ctx context.Context, teamName string) (*model.Team, error) {
	err := r.CheckExistsTeam(ctx, teamName)
	if err != nil {
		return nil, ErrTeamNotFound
	}
	var team model.Team
	team.Team = teamName

	rows, err := r.pool.Query(ctx, `
        SELECT user_id, username, is_active 
        FROM reviewers 
        WHERE team_name = $1
    `, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user model.User
		err := rows.Scan(&user.ID, &user.Name, &user.Status)
		if err != nil {
			return nil, err
		}
		team.Users = append(team.Users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &team, nil

}

func (r *TeamRepository) CheckExistsTeam(ctx context.Context, teamName string) error {
	var flg bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)
		`, teamName).Scan(&flg)
	if err != nil {
		return err
	}
	if !flg {
		return ErrTeamNotFound
	}
	return nil
}
