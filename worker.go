package nomadatc

import (
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/concourse/atc"
	"github.com/concourse/atc/creds"
	"github.com/concourse/atc/db"
	"github.com/concourse/atc/worker"
)

type WorkerClient struct {
}

func (w *WorkerClient) FindOrCreateContainer(lager.Logger, <-chan os.Signal, worker.ImageFetchingDelegate, db.ContainerOwner, db.ContainerMetadata, worker.ContainerSpec, creds.VersionedResourceTypes) (worker.Container, error) {
	panic("not implemented")
}

func (w *WorkerClient) FindContainerByHandle(lager.Logger, int, string) (worker.Container, bool, error) {
	panic("not implemented")
}

func (w *WorkerClient) LookupVolume(lager.Logger, string) (worker.Volume, bool, error) {
	panic("not implemented")
}

func (w *WorkerClient) FindResourceTypeByPath(path string) (atc.WorkerResourceType, bool) {
	panic("not implemented")
}

func (w *WorkerClient) Satisfying(lager.Logger, worker.WorkerSpec, creds.VersionedResourceTypes) (worker.Worker, error) {
	panic("not implemented")
}

func (w *WorkerClient) AllSatisfying(lager.Logger, worker.WorkerSpec, creds.VersionedResourceTypes) ([]worker.Worker, error) {
	panic("not implemented")
}

func (w *WorkerClient) RunningWorkers(lager.Logger) ([]worker.Worker, error) {
	panic("not implemented")
}

var _ = worker.Client(&WorkerClient{})
