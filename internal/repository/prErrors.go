package repository

import "errors"

var (
	ErrPullRequestAlreadyExists = errors.New("PR already exists")
	ErrAuthorNotFound           = errors.New("Author not found")
	ErrAuthorIsNotActive        = errors.New("Author is not active")
	ErrNotReviewersInTeam       = errors.New("Not reviewers in team")
	ErrPullRequestNotFound      = errors.New("Pull request not found")
	ErrPullRequestAlreadyMerged = errors.New("Pull request already merged")
	ErrReviewerIsNotAssigned    = errors.New("reviewer is not assigned to this PR")
)
