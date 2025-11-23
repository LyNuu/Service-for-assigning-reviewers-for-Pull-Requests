package service

import (
	"avitoMerchStore/internal/model"
	"context"
)

type teamRepository interface {
	AddTeam(ctx context.Context, team *model.Team) (*model.Team, error)
	GetTeamByName(ctx context.Context, teamName string) (*model.Team, error)
}
