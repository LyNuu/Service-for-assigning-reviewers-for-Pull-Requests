package handler

import (
	"avitoMerchStore/internal/model"
	"context"
)

type prService interface {
	CreatePR(ctx context.Context, pr *model.PullRequest) (*model.PullRequest, error)
	MergePR(ctx context.Context, prID string) (*model.PullRequest, error)
	ReassignPR(ctx context.Context, pr *model.PullRequest, oldId string) (*model.PullRequest, error)
}
