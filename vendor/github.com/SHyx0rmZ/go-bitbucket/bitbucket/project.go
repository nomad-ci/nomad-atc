package bitbucket

type Project interface {
	GetKey() string
	Repositories() ([]Repository, error)
}
