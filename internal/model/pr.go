package model

type PRStatus string

const (
	StatusOpen   PRStatus = "OPEN"
	StatusMerged PRStatus = "MERGED"
)

type PullRequest struct {
	ID        string
	Name      string
	AuthorID  string
	Status    PRStatus
	Reviewers []string
}
