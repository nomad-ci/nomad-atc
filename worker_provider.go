package nomadatc

import (
	"os"
	"time"

	c "code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/lager"
	"github.com/concourse/atc"
	"github.com/concourse/atc/creds"
	"github.com/concourse/atc/db"
	"github.com/concourse/atc/worker"
	"github.com/concourse/baggageclaim"
	"github.com/cppforlife/go-semi-semantic/version"
)

type WorkerProvider struct {
	Logger                lager.Logger
	WorkerResourceFactory db.WorkerBaseResourceTypeFactory
	WorkerFactory         db.WorkerFactory
	TeamFactory           db.TeamFactory
	Clock                 c.Clock
	Driver                *Driver
}

func (w *WorkerProvider) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	var resourceTypes []atc.WorkerResourceType
	resourceTypes = append(resourceTypes, atc.WorkerResourceType{
		Type:  "git",
		Image: "concourse/git-resource",
	})

	workerInfo := atc.Worker{
		ActiveContainers: 0,
		ResourceTypes:    resourceTypes,
		Platform:         "linux",
		Tags:             []string{},
		Name:             "nomad",
	}

	_, err := w.WorkerFactory.SaveWorker(workerInfo, 30*time.Second)
	if err != nil {
		w.Logger.Error("could-not-save-nomad-worker-provided", err)
		return err
	}

	ticker := w.Clock.NewTicker(10 * time.Second)

	close(ready)

dance:
	for {
		select {
		case <-ticker.C():
			_, err = w.WorkerFactory.SaveWorker(workerInfo, 30*time.Second)
			if err != nil {
				w.Logger.Error("could-not-save-nomad-worker-provided", err)
			}
		case <-signals:
			ticker.Stop()
			break dance
		}
	}

	return nil
}

func (w *WorkerProvider) Register(p *Process) int64 {
	return w.Driver.Register(p)
}

func (w *WorkerProvider) Deregister(s int64) {
	w.Driver.Deregister(s)
}

func (w *WorkerProvider) RunningWorkers(logger lager.Logger) ([]worker.Worker, error) {
	w.Logger.Debug("nomad-running-workers")

	savedWorkers, err := w.WorkerFactory.Workers()
	if err != nil {
		return nil, err
	}

	tikTok := c.NewClock()

	workers := []worker.Worker{}

	for _, savedWorker := range savedWorkers {
		if savedWorker.State() != db.WorkerStateRunning {
			continue
		}

		ww := w.NewGardenWorker(logger, tikTok, savedWorker)

		workers = append(workers, ww)
	}

	return workers, nil
}

func (w *WorkerProvider) FindWorkerForContainer(logger lager.Logger, teamID int, handle string) (worker.Worker, bool, error) {
	w.Logger.Debug("nomad-find-worker-for-container")
	return nil, false, nil
}

func (w *WorkerProvider) FindWorkerForContainerByOwner(logger lager.Logger, teamID int, owner db.ContainerOwner) (worker.Worker, bool, error) {
	w.Logger.Debug("nomad-find-worker-for-container-by-owner")
	return nil, false, nil
}

func (w *WorkerProvider) NewGardenWorker(logger lager.Logger, tikTok c.Clock, savedWorker db.Worker) worker.Worker {
	w.Logger.Debug("nomad-new-garden-worker")
	ww := &Worker{w.Logger, w, savedWorker}
	return ww
}

var _ = worker.WorkerProvider(&WorkerProvider{})

type Worker struct {
	Logger   lager.Logger
	Provider *WorkerProvider

	dbWorker db.Worker
}

func (w *Worker) FindOrCreateContainer(logger lager.Logger, signals <-chan os.Signal, images worker.ImageFetchingDelegate, owner db.ContainerOwner, md db.ContainerMetadata, spec worker.ContainerSpec, vrt creds.VersionedResourceTypes) (worker.Container, error) {
	cont := &Container{
		Worker:  w,
		Logger:  logger,
		signals: signals,
		spec:    spec,
		md:      md,
	}

	for _, input := range spec.Inputs {
		volume := &Volume{
			Logger:    logger,
			Source:    input.Source(),
			Container: cont,
			path:      input.DestinationPath(),
		}

		mount := worker.VolumeMount{
			Volume:    volume,
			MountPath: input.DestinationPath(),
		}

		cont.inputs = append(cont.inputs, mount)
		cont.mounts = append(cont.mounts, mount)
	}

	for _, outputPath := range spec.Outputs {
		volume := &Volume{
			Logger:    logger,
			Container: cont,
			path:      outputPath,
		}

		mount := worker.VolumeMount{
			Volume:    volume,
			MountPath: outputPath,
		}

		cont.mounts = append(cont.mounts, mount)
	}

	logger.Debug("nomad-find-or-create-container")
	return cont, nil
}

func (w *Worker) FindContainerByHandle(lager.Logger, int, string) (worker.Container, bool, error) {
	panic("not implemented")
}

func (w *Worker) LookupVolume(lager.Logger, string) (worker.Volume, bool, error) {
	panic("not implemented")
}

func (w *Worker) FindResourceTypeByPath(path string) (atc.WorkerResourceType, bool) {
	panic("not implemented")
}

func (w *Worker) Satisfying(lager.Logger, worker.WorkerSpec, creds.VersionedResourceTypes) (worker.Worker, error) {
	return w, nil
}

func (w *Worker) AllSatisfying(lager.Logger, worker.WorkerSpec, creds.VersionedResourceTypes) ([]worker.Worker, error) {
	return []worker.Worker{w}, nil
}

func (w *Worker) RunningWorkers(lager.Logger) ([]worker.Worker, error) {
	return []worker.Worker{w}, nil
}

func (w *Worker) ActiveContainers() int {
	panic("not implemented")
}

func (w *Worker) Description() string {
	panic("not implemented")
}

func (w *Worker) Name() string {
	return "nomad"
}

func (w *Worker) ResourceTypes() []atc.WorkerResourceType {
	panic("not implemented")
}

func (w *Worker) Tags() atc.Tags {
	panic("not implemented")
}

func (w *Worker) Uptime() time.Duration {
	panic("not implemented")
}

func (w *Worker) IsOwnedByTeam() bool {
	return true
}

func (w *Worker) IsVersionCompatible(lager.Logger, *version.Version) bool {
	panic("not implemented")
}

func (w *Worker) FindVolumeForResourceCache(logger lager.Logger, resourceCache *db.UsedResourceCache) (worker.Volume, bool, error) {
	return nil, false, nil
}

func (w *Worker) FindVolumeForTaskCache(lager.Logger, int, int, string, string) (worker.Volume, bool, error) {
	panic("not implemented")
}

func (w *Worker) GardenClient() garden.Client {
	panic("not implemented")
}

func (w *Worker) BaggageclaimClient() baggageclaim.Client {
	panic("not implemented")
}

func (w *Worker) EnsureCertsVolumeExists(logger lager.Logger) error {
	panic("not implemented")
}

var _ = worker.Worker(&Worker{})
