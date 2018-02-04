package bitbucket

type Auth interface{}

type BasicAuth struct {
	Username string
	Password string
}

var BitbucketAuth contextKeyBitbucketAuth

type contextKeyBitbucketAuth struct{}
