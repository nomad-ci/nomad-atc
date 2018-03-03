package nomadatc

import (
	fmt "fmt"
	io "io"

	"code.cloudfoundry.org/lager"
	"github.com/concourse/atc/db"
	"github.com/concourse/atc/worker"
	"github.com/concourse/baggageclaim"
)

type Volume struct {
	Logger  lager.Logger
	Source  worker.ArtifactSource
	buildId int
	Driver  *Driver

	handle string

	path string
	priv bool

	props baggageclaim.VolumeProperties
}

func (v *Volume) Handle() string {
	return v.handle
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
	v.Logger.Info("nomad-volume-stream-out", lager.Data{"path": path})
	return v.Driver.StreamBuildVolume(v.buildId, v.handle, path)
}

func (v *Volume) COWStrategy() baggageclaim.COWStrategy {
	return baggageclaim.COWStrategy{}
}

func (v *Volume) InitializeResourceCache(cache *db.UsedResourceCache) error {
	v.Logger.Debug("nomad-volume-init-resource-cache-STUB")
	return nil
}

func (v *Volume) InitializeTaskCache(lager.Logger, int, string, string, bool) error {
	v.Logger.Debug("nomad-initialize-task-cache-STUB")
	return nil
}

func (v *Volume) CreateChildForContainer(db.CreatingContainer, string) (db.CreatingVolume, error) {
	v.Logger.Debug("nomad-initialize-create-child-for-conatiner-STUB")
	return nil, fmt.Errorf("not implemented")
}

func (v *Volume) Destroy() error {
	return nil
}
