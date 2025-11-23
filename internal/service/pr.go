package service

import (
	"avitoMerchStore/internal/model"
	"context"
	"errors"
)

type PrService struct {
	repository PrRepository
}

func NewPrService(repository PrRepository) *PrService {
	return &PrService{repository: repository}
}

func (s *PrService) CreatePR(ctx context.Context, pr *model.PullRequest) (*model.PullRequest, error) {
	p, err := s.repository.CreatePR(ctx, pr)
	if err != nil {
		if errors.Is(err, ErrAuthorNotFound) {
			return nil, ErrAuthorNotFound
		}
		if errors.Is(err, ErrTeamNotFound) {
			return nil, ErrTeamNotFound
		}
		return nil, err
	}
	return p, nil
}

func (s *PrService) MergePR(ctx context.Context, prID string) (*model.PullRequest, error) {
	p, err := s.repository.MergePR(ctx, prID)
	if err != nil {
		if errors.Is(err, ErrPullRequestNotFound) {
			return nil, ErrPullRequestNotFound
		}
		return nil, err
	}
	return p, nil
}

func (s *PrService) ReassignPR(ctx context.Context, pr *model.PullRequest, oldId string) (*model.PullRequest, error) {
	p, err := s.repository.ReassignPR(ctx, pr, oldId)
	if err != nil {
		if errors.Is(err, ErrPullRequestNotFound) {
			return nil, ErrPullRequestNotFound
		}
		if errors.Is(err, ErrPullRequestAlreadyMerged) {
			return nil, ErrPullRequestAlreadyMerged
		}
		return nil, err
	}
	return p, nil
}
