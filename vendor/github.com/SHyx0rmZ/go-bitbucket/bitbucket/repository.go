package bitbucket

type Repository interface {
	PullRequest(id int) (PullRequest, error)
	PullRequests() ([]PullRequest, error)
	CreatePullRequest(from Repository, fromID string, toID string, title string, description string, reviewers ...string) (PullRequest, error)

	GetProject() Project
	GetName() string
	GetSlug() string
}
