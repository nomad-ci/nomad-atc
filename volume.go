package nomadatc

import (
	io "io"
	"path/filepath"

	"code.cloudfoundry.org/lager"
	"github.com/concourse/atc/db"
	"github.com/concourse/atc/worker"
	"github.com/concourse/baggageclaim"
)

type Volume struct {
	Logger    lager.Logger
	Source    worker.ArtifactSource
	Container *Container

	path string
	priv bool

	props baggageclaim.VolumeProperties
}

func (v *Volume) Handle() string {
	panic("not implemented")
}

func (v *Volume) Path() string {
	return v.path
}

func (v *Volume) SetProperty(key string, value string) error {
	v.Logger.Debug("nomad-volume-set-property", lager.Data{"key": key, "value": value})

	if v.props == nil {
		v.props = make(baggageclaim.VolumeProperties)
	}

	v.props[key] = value

	return nil
}

func (v *Volume) Properties() (baggageclaim.VolumeProperties, error) {
	return v.props, nil
}

func (v *Volume) SetPrivileged(b bool) error {
	v.priv = b
	return nil
}

func (v *Volume) StreamIn(path string, tarStream io.Reader) error {
	panic("not implemented")
}

func (v *Volume) StreamOut(path string) (io.ReadCloser, error) {
	v.Logger.Debug("nomad-volume-stream-out", lager.Data{"path": path})
	return v.Container.process.StreamFile(filepath.Join(v.path, path))
}

func (v *Volume) COWStrategy() baggageclaim.COWStrategy {
	panic("not implemented")
}

func (v *Volume) InitializeResourceCache(cache *db.UsedResourceCache) error {
	v.Logger.Debug("nomad-volume-init-resource-cache-STUB")
	return nil
}

func (v *Volume) InitializeTaskCache(lager.Logger, int, string, string, bool) error {
	panic("not implemented")
}

func (v *Volume) CreateChildForContainer(db.CreatingContainer, string) (db.CreatingVolume, error) {
	panic("not implemented")
}

func (v *Volume) Destroy() error {
	panic("not implemented")
}
