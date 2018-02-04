package bitbucket

type PullRequest interface {
	GetID() int
	GetVersion() int
	GetTitle() string
	GetDescription() string
	GetState() string
	GetAuthorName() string
	GetFromRef() string
	GetToRef() string
}
