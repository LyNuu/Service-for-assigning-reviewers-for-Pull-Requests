package handler

import (
	"avitoMerchStore/internal/model"
	"context"
)

type userService interface {
	SetIsActive(ctx context.Context, id string, status bool) (*model.User, error)
	GetPrById(ctx context.Context, id string) (*[]model.PullRequest, error)
}
