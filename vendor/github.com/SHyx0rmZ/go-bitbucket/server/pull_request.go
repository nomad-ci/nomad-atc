package server

type pullRequest struct {
	client *Client

	ID          int    `json:"id"`
	Version     int    `json:"version"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"`
	Open        bool   `json:"open"`
	Closed      bool   `json:"closed"`
	CreatedDate int    `json:"createdDate"`
	UpdatedDate int    `json:"updatedDate"`
	FromRef     struct {
		ID         string `json:"id"`
		Repository struct {
			Slug    string `json:"slug"`
			Name    string `json:"name"`
			Project struct {
				Key string `json:"key"`
			} `json:"project"`
		} `json:"repository"`
	} `json:"fromRef"`
	ToRef struct {
		ID         string `json:"id"`
		Repository struct {
			Slug    string `json:"slug"`
			Name    string `json:"name"`
			Project struct {
				Key string `json:"key"`
			} `json:"project"`
		} `json:"repository"`
	} `json:"toRef"`
	Locked bool `json:"locked"`
	Author struct {
		User     user   `json:"user"`
		Role     string `json:"role"`
		Approved bool   `json:"approved"`
		Status   string `json:"status"`
	}
	Reviewers    []reviewer `json:"reviewers"`
	Participants []struct {
		User     user   `json:"user"`
		Role     string `json:"role"`
		Approved bool   `json:"approved"`
		Status   string `json:"status"`
	}
}

type reviewer struct {
	User               user   `json:"user"`
	LastReviewedCommit string `json:"lastReviewedCommit"`
	Role               string `json:"role"`
	Approved           bool   `json:"approved"`
	Status             string `json:"status"`
}

func (p *pullRequest) SetClient(c *Client) {
	p.client = c
	p.Author.User.SetClient(c)

	for index := range p.Reviewers {
		p.Reviewers[index].User.SetClient(c)
	}

	for index := range p.Participants {
		p.Participants[index].User.SetClient(c)
	}
}

func (p pullRequest) GetID() int {
	return p.ID
}

func (p pullRequest) GetVersion() int {
	return p.Version
}

func (p pullRequest) GetTitle() string {
	return p.Title
}

func (p pullRequest) GetDescription() string {
	return p.Description
}

func (p pullRequest) GetState() string {
	return p.State
}

func (p pullRequest) GetAuthorName() string {
	return p.Author.User.Name
}

func (p pullRequest) GetFromRef() string {
	return p.FromRef.ID
}

func (p pullRequest) GetToRef() string {
	return p.ToRef.ID
}
