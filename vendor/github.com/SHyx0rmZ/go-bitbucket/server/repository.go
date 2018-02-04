package server

import (
	"github.com/SHyx0rmZ/go-bitbucket/bitbucket"
	"strconv"
)

type repository struct {
	client *Client

	Slug                      string  `json:"slug"`
	ID                        int     `json:"id"`
	Name                      string  `json:"name"`
	SourceControlManagementID string  `json:"scmId"`
	State                     string  `json:"state"`
	StatusMessage             string  `json:"statusMessage"`
	Forkable                  bool    `json:"forkable"`
	Project                   project `json:"project"`
	Public                    bool    `json:"public"`
}

func (r *repository) GetProject() bitbucket.Project {
	return &r.Project
}

func (r *repository) GetName() string {
	return r.Name
}

func (r *repository) GetSlug() string {
	return r.Slug
}

func (r *repository) SetClient(c *Client) {
	r.client = c
	r.Project.SetClient(c)
}

func (r *repository) PullRequest(id int) (bitbucket.PullRequest, error) {
	var pullRequest pullRequest

	err := r.client.request("/rest/api/1.0/projects/"+r.Project.Key+"/repos/"+r.Slug+"/pull-requests/"+strconv.Itoa(id), &pullRequest)
	if err != nil {
		return nil, err
	}

	return &pullRequest, nil
}

func (r *repository) PullRequests() ([]bitbucket.PullRequest, error) {
	pullRequests := make([]pullRequest, 0, 0)

	err := r.client.pagedRequest("/rest/api/1.0/projects/"+r.Project.Key+"/repos/"+r.Slug+"/pull-requests", &pullRequests)
	if err != nil {
		return nil, err
	}

	bitbucketPullRequests := make([]bitbucket.PullRequest, len(pullRequests))
	for index := range pullRequests {
		bitbucketPullRequests[index] = &pullRequests[index]
	}

	return bitbucketPullRequests, nil
}

func (r *repository) CreatePullRequest(from bitbucket.Repository, fromID string, toID string, title string, description string, reviewers ...string) (bitbucket.PullRequest, error) {
	pr := pullRequest{
		Title:       title,
		Description: description,
		State:       "OPEN",
		Open:        true,
		Closed:      false,
		Locked:      false,
	}

	if from == nil {
		from = r
	}

	pr.FromRef.ID = fromID
	pr.FromRef.Repository.Project.Key = from.GetProject().GetKey()
	pr.FromRef.Repository.Name = from.GetName()
	pr.FromRef.Repository.Slug = from.GetSlug()
	pr.ToRef.ID = toID
	pr.ToRef.Repository.Project.Key = r.Project.Key
	pr.ToRef.Repository.Name = r.Name
	pr.ToRef.Repository.Slug = r.Slug

	for _, user := range reviewers {
		reviewer := reviewer{}
		reviewer.User.Name = user

		pr.Reviewers = append(pr.Reviewers, reviewer)
	}

	var fullPullRequest pullRequest

	err := r.client.requestPost("/rest/api/1.0/projects/"+r.Project.Key+"/repos/"+r.Slug+"/pull-requests", &fullPullRequest, pr)
	if err != nil {
		return nil, err
	}

	return fullPullRequest, nil
}
