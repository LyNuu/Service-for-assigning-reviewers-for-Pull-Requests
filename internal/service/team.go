package service

import (
	"avitoMerchStore/internal/model"
	"context"
	"errors"
)

type TeamService struct {
	repository teamRepository
}

func NewTeamService(repository teamRepository) *TeamService {
	return &TeamService{repository: repository}
}

func (s *TeamService) AddTeam(ctx context.Context, team *model.Team) (*model.Team, error) {
	_, err := s.repository.AddTeam(ctx, team)
	if err != nil {
		if errors.Is(err, ErrTeamAlreadyExists) {
			return nil, ErrTeamAlreadyExists
		}
		return nil, err
	}
	return team, nil
}

func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*model.Team, error) {
	team, err := s.repository.GetTeamByName(ctx, teamName)
	if err != nil {
		if errors.Is(err, ErrTeamNotFound) {
			return nil, ErrTeamNotFound
		}
		return nil, err
	}
	return team, nil
}
