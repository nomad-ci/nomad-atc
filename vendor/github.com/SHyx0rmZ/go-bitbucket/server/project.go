package server

import "github.com/SHyx0rmZ/go-bitbucket/bitbucket"

type project struct {
	client *Client

	Key         string `json:"key"`
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Public      bool   `json:"public"`
	Type        string `json:"type"`
}

func (p *project) SetClient(c *Client) {
	p.client = c
}

func (p project) GetKey() string {
	return p.Key
}

func (p *project) Repositories() ([]bitbucket.Repository, error) {
	repositories := make([]repository, 0, 0)

	err := p.client.pagedRequest("/rest/api/1.0/projects/"+p.Key+"/repos", &repositories)
	if err != nil {
		return nil, err
	}

	bitbucketRepositories := make([]bitbucket.Repository, len(repositories))

	for i := range repositories {
		bitbucketRepositories[i] = &repositories[i]
	}

	return bitbucketRepositories, nil
}
