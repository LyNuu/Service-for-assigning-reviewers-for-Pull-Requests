package service

import "errors"

var (
	ErrTeamAlreadyExists = errors.New("Team already exists")
	ErrTeamNotFound      = errors.New("Команда не найдена")
)
