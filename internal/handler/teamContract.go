package handler

import (
	"avitoMerchStore/internal/model"
	"context"
)

type teamService interface {
	AddTeam(ctx context.Context, team *model.Team) (*model.Team, error)
	GetTeam(ctx context.Context, teamName string) (*model.Team, error)
}
