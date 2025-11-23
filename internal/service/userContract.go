package service

import (
	"avitoMerchStore/internal/model"
	"context"
)

type userRepository interface {
	GetById(ctx context.Context, id string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	GetPrById(ctx context.Context, id string) (*[]model.PullRequest, error)
}
