package cloud

import "time"

type mergeEndpoint struct {
	Commit struct {
		Hash string `json:"hash"`
	} `json:"commit"`
	Repository repository `json:"repository"`
}

type pullrequest struct {
	client *Client

	Id          int           `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	CreatedDate time.Time     `json:"created_on"`
	UpdatedDate time.Time     `json:"updated_on"`
	State       string        `json:"state"`
	Author      user          `json:"author"`
	Source      mergeEndpoint `json:"source"`
	Destination mergeEndpoint `json:"destination"`
}

func (pr *pullrequest) GetID() int {
	return pr.Id
}

func (pr *pullrequest) GetVersion() int {
	// no version, maybe unix epoch of updatedDate?
	return int(pr.UpdatedDate.Unix())
}

func (pr *pullrequest) GetTitle() string {
	return pr.Title
}

func (pr *pullrequest) GetDescription() string {
	return pr.Description
}

func (pr *pullrequest) GetState() string {
	return pr.State
}

func (pr *pullrequest) GetAuthorName() string {
	return pr.Author.GetName()
}

func (pr *pullrequest) GetFromRef() string {
	return pr.Source.Commit.Hash
}

func (pr *pullrequest) GetToRef() string {
	return pr.Destination.Commit.Hash
}

func (pr *pullrequest) SetClient(c *Client) {
	pr.client = c
}
