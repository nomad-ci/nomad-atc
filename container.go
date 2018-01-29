package nomadatc

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/lager"
	"github.com/concourse/atc/db"
	"github.com/concourse/atc/worker"
	nomad "github.com/hashicorp/nomad/api"
	context "golang.org/x/net/context"
)

const (
	taskProcessID              = "task"
	taskProcessPropertyName    = "concourse:task-process"
	taskExitStatusPropertyName = "concourse:exit-status"
)

type Container struct {
	Worker *Worker
	Logger lager.Logger

	id      string
	signals <-chan os.Signal
	spec    worker.ContainerSpec
	inputs  []worker.VolumeMount
	mounts  []worker.VolumeMount
	props   map[string]string
	md      db.ContainerMetadata

	process *Process
}

func (c *Container) Handle() string {
	panic("not implemented")
}

func (c *Container) Stop(kill bool) error {
	if c.id == "" {
		return nil
	}

	cfg := c.Worker.Provider.nomadConfig()

	n, err := nomad.NewClient(cfg)
	if err != nil {
		return err
	}

	_, _, err = n.Jobs().Deregister(c.id, false, nil)
	return err
}

func (c *Container) Info() (garden.ContainerInfo, error) {
	panic("not implemented")
}

func (c *Container) StreamIn(spec garden.StreamInSpec) error {
	panic("not implemented")
}

func (c *Container) StreamOut(spec garden.StreamOutSpec) (io.ReadCloser, error) {
	panic("not implemented")
}

func (c *Container) CurrentBandwidthLimits() (garden.BandwidthLimits, error) {
	panic("not implemented")
}

func (c *Container) CurrentCPULimits() (garden.CPULimits, error) {
	panic("not implemented")
}

func (c *Container) CurrentDiskLimits() (garden.DiskLimits, error) {
	panic("not implemented")
}

func (c *Container) CurrentMemoryLimits() (garden.MemoryLimits, error) {
	panic("not implemented")
}

func (c *Container) NetIn(hostPort uint32, containerPort uint32) (uint32, uint32, error) {
	panic("not implemented")
}

func (c *Container) NetOut(netOutRule garden.NetOutRule) error {
	panic("not implemented")
}

func (c *Container) BulkNetOut(netOutRules []garden.NetOutRule) error {
	panic("not implemented")
}

func StringToPtr(s string) *string {
	return &s
}

func IntToPtr(i int) *int {
	return &i
}

func (c *Container) Run(spec garden.ProcessSpec, io garden.ProcessIO) (garden.Process, error) {
	cfg := c.Worker.Provider.nomadConfig()

	n, err := nomad.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	proc := &Process{
		Container: c,
		logger:    c.Logger,
		id:        spec.Path,
		api:       n,
		spec:      spec,
		io:        io,
		done:      make(chan struct{}, 1),
		requests:  make(chan *StreamFileRequest, 1),
		buildId:   c.md.BuildID,
	}

	stream := c.Worker.Provider.Register(proc)

	jobId := fmt.Sprintf("atc:%x:%s:%s:%s:%s", stream, c.md.Type, c.md.PipelineName, c.md.JobName, c.md.StepName)

	proc.job = jobId
	c.id = jobId

	c.process = proc

	var tc TaskConfig
	tc.Stream = stream
	tc.Host = fmt.Sprintf("%s:%d", c.Worker.Provider.InternalIP, c.Worker.Provider.InternalPort)
	tc.Path = spec.Path
	tc.Args = spec.Args
	tc.Dir = spec.Dir
	tc.Env = spec.Env

	if c.md.Type == db.ContainerTypeGet ||
		(c.md.Type == db.ContainerTypeTask && len(c.spec.Outputs) > 0) {
		tc.WaitForVolumes = true
	}

	for _, m := range c.inputs {
		tc.Inputs = append(tc.Inputs, m.MountPath)
	}

	var (
		dockerImage string
		priv        bool
	)

	ir := c.spec.ImageSpec.ImageResource
	if ir == nil || ir.Type != "docker-image" {
		if br, ok := AllResources[c.spec.ImageSpec.ResourceType]; ok {
			dockerImage = br.Image
			if br.Version != "" {
				dockerImage = dockerImage + ":" + br.Version
			}

			priv = br.Privileged
		} else {
			fmt.Errorf("unsupported resource: %s", c.spec.ImageSpec.ResourceType)
		}
	} else {
		src, err := ir.Source.Evaluate()
		if err != nil {
			return nil, err
		}

		if image, ok := src["repository"]; ok {
			dockerImage = fmt.Sprintf("%s", image)
		} else {
			return nil, errors.New("no docker image specified")
		}

		if tag, ok := src["tag"]; ok {
			dockerImage = fmt.Sprintf("%s:%s", dockerImage, tag)
		}
	}

	config, err := json.Marshal(&tc)
	if err != nil {
		return nil, err
	}

	var tmpl nomad.Template
	tmpl.ChangeMode = StringToPtr("noop")
	tmpl.EmbeddedTmpl = StringToPtr(string(config))
	tmpl.DestPath = StringToPtr("local/config.json")

	job := &nomad.Job{
		ID:   StringToPtr(jobId),
		Name: StringToPtr(jobId),

		Datacenters: c.Worker.Provider.Datacenters,

		Type: StringToPtr("batch"),

		TaskGroups: []*nomad.TaskGroup{
			&nomad.TaskGroup{
				Name: StringToPtr("atc"),
				RestartPolicy: &nomad.RestartPolicy{
					Attempts: IntToPtr(0),
					Mode:     StringToPtr("fail"),
				},

				Meta: map[string]string{
					"config": "/local/config.json",
				},

				Tasks: []*nomad.Task{
					&nomad.Task{
						Name:      "atc",
						Driver:    "docker",
						Templates: []*nomad.Template{&tmpl},
						Artifacts: []*nomad.TaskArtifact{
							&nomad.TaskArtifact{
								GetterSource: StringToPtr(
									fmt.Sprintf("http://%s:%d/_driver.tar.gz",
										c.Worker.Provider.InternalIP, c.Worker.Provider.InternalPort+1)),
								RelativeDest: StringToPtr("local/driver"),
							},
						},
						Config: map[string]interface{}{
							"image":      dockerImage,
							"command":    "/local/driver/driver",
							"privileged": priv,
							"mounts": []map[string]interface{}{
								{
									"target": "/scratch",
									"volume_options": []map[string]interface{}{
										{
											"driver_options": []map[string]interface{}{
												{
													"name": "local",
												},
											},
										},
									},
								},
							},
						},
						Resources: &nomad.Resources{
							CPU:      IntToPtr(500),
							MemoryMB: IntToPtr(256),
						},
					},
				},
			},
		},
	}

	resp, _, err := n.Jobs().Register(job, nil)
	if err != nil {
		c.Logger.Error("nomad-register-job-failed", err)
		return nil, err
	}

	c.Logger.Info("nomad-registered-job", lager.Data{"job": *job.ID, "eval": resp.EvalID, "image": dockerImage})
	go func() {
		proc.monitor(resp.EvalID)
	}()

	return proc, nil
}

var ErrUnknownProcess = errors.New("unknown process")

func (c *Container) Attach(processID string, io garden.ProcessIO) (garden.Process, error) {
	c.Logger.Debug("nomad-container-attach")
	return nil, ErrUnknownProcess
}

func (c *Container) Metrics() (garden.Metrics, error) {
	panic("not implemented")
}

func (c *Container) SetGraceTime(graceTime time.Duration) error {
	panic("not implemented")
}

func (c *Container) Properties() (garden.Properties, error) {
	panic("not implemented")
}

var ErrNotSet = errors.New("not set")

func (c *Container) Property(name string) (string, error) {
	c.Logger.Debug("nomad-container-property", lager.Data{"name": name})

	prop, ok := c.props[name]
	if ok {
		return prop, nil
	}

	return "", ErrNotSet
}

func (c *Container) SetProperty(name string, value string) error {
	if c.props == nil {
		c.props = make(map[string]string)
	}

	c.Logger.Debug("nomad-container-set-property", lager.Data{
		"name":  name,
		"value": value,
	})

	c.props[name] = value

	return nil
}

func (c *Container) RemoveProperty(name string) error {
	panic("not implemented")
}

func (c *Container) Destroy() error {
	panic("not implemented")
}

func (c *Container) VolumeMounts() []worker.VolumeMount {
	return c.mounts
}

func (c *Container) WorkerName() string {
	panic("not implemented")
}

func (c *Container) MarkAsHijacked() error {
	panic("not implemented")
}

var _ = worker.Container(&Container{})

type Process struct {
	Container *Container
	logger    lager.Logger
	id        string
	api       *nomad.Client
	job       string
	spec      garden.ProcessSpec
	io        garden.ProcessIO

	status int
	done   chan struct{}

	requests chan *StreamFileRequest

	buildId int
	cancel  func()
}

func (p *Process) monitor(evalid string) {
	n := p.api

	var q nomad.QueryOptions
	q.WaitTime = 10 * time.Second

	var (
		list []*nomad.AllocationListStub
		meta *nomad.QueryMeta
		err  error
		id   string
	)

outer:
	for {
		list, meta, err = n.Jobs().Allocations(p.job, true, &q)
		if err != nil {
			break
		}

		for _, alloc := range list {
			if alloc.EvalID == evalid {
				id = alloc.ID
				break outer
			}
		}

		q.WaitIndex = meta.LastIndex
	}

	if id == "" {
		p.logger.Error("nomad-process-allocation-failed", err, lager.Data{"job": p.job})
		p.status = -1
		close(p.done)
		return
	}

	p.logger.Info("nomad-allocation-started", lager.Data{"alloc": id})

	q.WaitIndex = meta.LastIndex
	q.WaitTime = 10 * time.Second

	var alloc *nomad.Allocation

	var seenRunning bool

	for {
		alloc, meta, err = n.Allocations().Info(id, &q)
		if err != nil {
			p.logger.Error("nomad-process-allocation-info-failed", err)
			p.status = -1
			close(p.done)
			return
		}

		p.logger.Debug("nomad-process-alloc-status", lager.Data{
			"alloc":  alloc.ID,
			"status": alloc.ClientStatus,
		})

		switch alloc.ClientStatus {
		case "running":
			if !seenRunning {
				p.logger.Info("nomad-allocation-running", lager.Data{"alloc": id, "job": alloc.JobID})
				seenRunning = true
			}

		case "complete", "failed":
			if alloc.ClientStatus == "failed" {
				p.logger.Error("nomad-allocation-failed", errors.New("allocation failed"), lager.Data{"job": p.job})
			} else {
				p.logger.Info("nomad-allocation-finished", lager.Data{"job": p.job, "exit": p.status})
			}
			// close(p.done)
			return
		}

		q.WaitIndex = meta.LastIndex
	}
}

func (p *Process) setFinished(code int32) {
	p.status = int(code)
	close(p.done)
}

type StreamFileRequest struct {
	Logger lager.Logger
	Path   string
	writer io.WriteCloser
}

func (fr *StreamFileRequest) Data(b []byte) error {
	_, err := fr.writer.Write(b)
	return err
}

func (fr *StreamFileRequest) Close() error {
	return fr.writer.Close()
}

type TransparentGzipReader struct {
	Under io.ReadCloser

	gzr *gzip.Reader
}

func (t *TransparentGzipReader) Read(b []byte) (int, error) {
	if t.gzr == nil {
		gzr, err := gzip.NewReader(t.Under)
		if err != nil {
			return 0, err
		}

		t.gzr = gzr
	}

	return t.gzr.Read(b)
}

func (t *TransparentGzipReader) Close() error {
	if t.gzr != nil {
		return t.gzr.Close()
	} else {
		return t.Under.Close()
	}
}

func (p *Process) StreamOut(path string) (io.ReadCloser, error) {
	r, w := io.Pipe()

	req := &StreamFileRequest{
		Logger: p.logger,
		Path:   path,
		writer: w,
	}

	p.requests <- req

	return r, nil
}

func (p *Process) NextFileRequest(ctx context.Context) (*StreamFileRequest, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case req, ok := <-p.requests:
		if !ok {
			return nil, io.EOF
		}

		return req, nil
	}
}

func (p *Process) ID() string {
	return p.id
}

func (p *Process) Wait() (int, error) {
	p.logger.Debug("nomad-process-waiting")
	<-p.done
	p.logger.Info("nomad-process-finished", lager.Data{"job": p.job, "exit": p.status})
	return p.status, nil
}

func (p *Process) SetTTY(garden.TTYSpec) error {
	panic("not implemented")
}

func (p *Process) Signal(garden.Signal) error {
	_, _, err := p.api.Jobs().Deregister(p.job, false, nil)
	return err
}

var _ = garden.Process(&Process{})
