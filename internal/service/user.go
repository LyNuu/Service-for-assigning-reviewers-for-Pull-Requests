package service

import (
	"avitoMerchStore/internal/model"
	"context"
	"errors"
)

type UserService struct {
	repository userRepository
}

func NewUserService(repository userRepository) *UserService {
	return &UserService{repository: repository}
}

func (s *UserService) SetIsActive(ctx context.Context, id string, status bool) (*model.User, error) {
	user, err := s.repository.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	user.Status = status
	err = s.repository.Update(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetPrById(ctx context.Context, id string) (*[]model.PullRequest, error) {
	pr, err := s.repository.GetPrById(ctx, id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return pr, nil
}
