package cloud

import (
	"github.com/SHyx0rmZ/go-bitbucket/bitbucket"
	"time"
)

type team struct {
	c *Client

	Username    string    `json:"username"`
	Website     string    `json:"website"`
	DisplayName string    `json:"display_name"`
	UUID        string    `json:"uuid"`
	CreatedOn   time.Time `json:"created_on"`
	Location    string    `json:"location"`
	Type        string    `json:"type"`
}

func (t *team) GetName() string {
	return t.Username
}

func (t *team) Members() ([]bitbucket.User, error) {
	users := make([]user, 0, 0)

	err := t.c.pagedRequest("/2.0/teams/"+t.GetName()+"/members", &users)
	if err != nil {
		return nil, err
	}

	bitbucketUsers := make([]bitbucket.User, len(users))

	for i := range users {
		bitbucketUsers[i] = &users[i]
	}

	return bitbucketUsers, nil
}

func (t *team) SetClient(c *Client) {
	t.c = c
}
