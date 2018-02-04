package cloud

import (
	"github.com/SHyx0rmZ/go-bitbucket/bitbucket"
)

type repository struct {
	Name        string `json:"full_name"`
	DisplayName string `json:"name"`
	Type        string `json:"type"`

	c *Client
}

func (r *repository) SetClient(c *Client) {
	r.c = c
}

func (r *repository) GetName() string {
	return r.Name
}

func (r *repository) GetDisplayName() string {
	return r.DisplayName
}

func (r *repository) PullRequests() ([]bitbucket.PullRequest, error) {
	prs := make([]pullrequest, 0, 0)

	err := r.c.pagedRequest("/2.0/repositories/"+r.Name+"/pullrequests", &prs)
	if err != nil {
		return nil, err
	}

	bitbucketPRs := make([]bitbucket.PullRequest, len(prs))
	for index := range prs {
		bitbucketPRs[index] = &prs[index]
	}

	return bitbucketPRs, nil
}

func (r *repository) PullRequest(id int) (bitbucket.PullRequest, error) {
	panic("implement me")
}

func (r *repository) CreatePullRequest(from bitbucket.Repository, fromID string, toID string, title string, description string, reviewers ...string) (bitbucket.PullRequest, error) {
	panic("implement me")
}

func (r *repository) GetProject() bitbucket.Project {
	panic("implement me")
}

func (r *repository) GetSlug() string {
	panic("implement me")
}
