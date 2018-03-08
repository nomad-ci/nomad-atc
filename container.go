package nomadatc

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	strconv "strconv"
	"strings"
	"sync"
	"time"

	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/lager"
	"github.com/concourse/atc/db"
	"github.com/concourse/atc/worker"
	"github.com/davecgh/go-spew/spew"
	nomad "github.com/hashicorp/nomad/api"
	"github.com/nomad-ci/nomad-atc/config"
	"github.com/nomad-ci/nomad-atc/rpc"
	context "golang.org/x/net/context"
)

const (
	taskProcessID              = "task"
	taskProcessPropertyName    = "concourse:task-process"
	taskExitStatusPropertyName = "concourse:exit-status"
)

// Container represents a running docker container in the
// nomad cluster. It satisfies the worker.Container interface.
type Container struct {
	Worker *Worker
	Logger lager.Logger

	handle  string
	id      string
	signals <-chan os.Signal
	spec    worker.ContainerSpec
	inputs  []worker.VolumeMount
	mounts  []worker.VolumeMount
	props   garden.Properties
	md      db.ContainerMetadata

	image   *worker.ImageResource
	process *Process

	inputVolumes  []*config.Volume
	outputVolumes []*config.Volume
}

func (c *Container) Handle() string {
	return c.handle
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

// Doesn't appear to be used anymore
func (c *Container) Info() (garden.ContainerInfo, error) {
	return garden.ContainerInfo{}, nil
}

func (c *Container) StreamIn(spec garden.StreamInSpec) error {
	panic("not implemented")
}

func (c *Container) StreamOut(spec garden.StreamOutSpec) (io.ReadCloser, error) {
	panic("not implemented")
}

func (c *Container) CurrentBandwidthLimits() (garden.BandwidthLimits, error) {
	return garden.BandwidthLimits{}, nil
}

func (c *Container) CurrentCPULimits() (garden.CPULimits, error) {
	return garden.CPULimits{}, nil
}

func (c *Container) CurrentDiskLimits() (garden.DiskLimits, error) {
	return garden.DiskLimits{}, nil
}

func (c *Container) CurrentMemoryLimits() (garden.MemoryLimits, error) {
	return garden.MemoryLimits{}, nil
}

func (c *Container) NetIn(hostPort uint32, containerPort uint32) (uint32, uint32, error) {
	return 0, 0, nil
}

func (c *Container) NetOut(netOutRule garden.NetOutRule) error {
	return nil
}

func (c *Container) BulkNetOut(netOutRules []garden.NetOutRule) error {
	return nil
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
		actions:   make(chan *rpc.Actions, 1),
		buildId:   c.md.BuildID,
		status:    -1,
	}

	var mem, cpu int

	if c.md.Type == db.ContainerTypeTask {
		mem = c.Worker.Provider.TaskMemory
		cpu = c.Worker.Provider.TaskCPU
	} else {
		mem = c.Worker.Provider.ResourceMemory
		cpu = c.Worker.Provider.ResourceCPU
	}

	const nmEnv = "NOMAD_MEMORY="

	for _, ev := range c.spec.Env {
		if strings.HasPrefix(ev, nmEnv) {
			i, err := strconv.Atoi(ev[len(nmEnv):])
			if err == nil {
				mem = i
				c.Logger.Info("nomad-container-plan-memory-override", lager.Data{"megabytes": mem})
			} else {
				c.Logger.Error("nomad-container-bad-plan-memory-override", err)
			}
		}
	}

	const ncEnv = "NOMAD_CPU="

	for _, ev := range c.spec.Env {
		if strings.HasPrefix(ev, ncEnv) {
			i, err := strconv.Atoi(ev[len(ncEnv):])
			if err == nil {
				cpu = i
				c.Logger.Info("nomad-container-plan-cpu-override", lager.Data{"cpu": cpu})
			} else {
				c.Logger.Error("nomad-container-bad-plan-memory-override", err)
			}
		}
	}

	stream := c.Worker.Provider.Register(proc)

	jobId := fmt.Sprintf("atc:%x:%s:%s:%s:%s", stream, c.md.Type, c.md.PipelineName, c.md.JobName, c.md.StepName)

	proc.job = jobId
	c.id = jobId

	c.process = proc

	var tc config.TaskConfig
	tc.Stream = stream
	tc.Host = fmt.Sprintf("%s:%d", c.Worker.Provider.InternalIP, c.Worker.Provider.InternalPort)
	tc.Path = spec.Path
	tc.Args = spec.Args
	tc.Dir = spec.Dir
	tc.Env = c.spec.Env

	if spec.TTY != nil {
		tc.UseTTY = true

		if spec.TTY.WindowSize != nil {
			tc.WindowSize = &config.WindowSize{
				Columns: spec.TTY.WindowSize.Columns,
				Rows:    spec.TTY.WindowSize.Rows,
			}
		}
	}

	if c.md.Type == db.ContainerTypeGet ||
		(c.md.Type == db.ContainerTypeTask && len(c.spec.Outputs) > 0) {
		tc.WaitForVolumes = true
	}

	tc.Inputs = c.inputVolumes
	tc.Outputs = c.outputVolumes

	var (
		dockerImage string
		priv        = c.spec.ImageSpec.Privileged
	)

	ir := c.image
	if ir == nil || ir.Type != "docker-image" {
		if br, ok := AllResources[c.spec.ImageSpec.ResourceType]; ok {
			dockerImage = br.Image
			if br.Version != "" {
				dockerImage = dockerImage + ":" + br.Version
			}

			priv = br.Privileged

			if br.Memory != 0 {
				mem = br.Memory
			}

		} else {
			fmt.Printf("unsupported resource: %s", c.spec.ImageSpec.ResourceType)
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

	if dockerImage == "" {
		spew.Dump(c.spec.ImageSpec)
		var typ string

		if ir == nil {
			typ = "<unknown>"
		} else {
			typ = ir.Type
		}

		return nil, fmt.Errorf("Unable to derive docker image to use from spec: %s", typ)
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
							CPU:      IntToPtr(cpu),
							MemoryMB: IntToPtr(mem),
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
		proc.monitor(io)
	}()

	return proc, nil
}

var ErrUnknownProcess = errors.New("unknown process")

func (c *Container) Attach(processID string, io garden.ProcessIO) (garden.Process, error) {
	c.Logger.Debug("nomad-container-attach")
	return nil, ErrUnknownProcess
}

func (c *Container) Metrics() (garden.Metrics, error) {
	return garden.Metrics{}, nil
}

func (c *Container) SetGraceTime(graceTime time.Duration) error {
	return nil
}

func (c *Container) Properties() (garden.Properties, error) {
	return c.props, nil
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
		c.props = make(garden.Properties)
	}

	c.Logger.Debug("nomad-container-set-property", lager.Data{
		"name":  name,
		"value": value,
	})

	c.props[name] = value

	return nil
}

func (c *Container) RemoveProperty(name string) error {
	if _, ok := c.props[name]; ok {
		delete(c.props, name)
		return nil
	}

	return ErrNotSet
}

func (c *Container) Destroy() error {
	return nil
}

func (c *Container) VolumeMounts() []worker.VolumeMount {
	return c.mounts
}

func (c *Container) WorkerName() string {
	return c.Worker.Name()
}

func (c *Container) MarkAsHijacked() error {
	return nil
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

	mu       sync.Mutex
	finished bool
	status   int
	done     chan struct{}

	requests chan *StreamFileRequest
	actions  chan *rpc.Actions

	buildId int
	cancel  func()
}

const maxAllocDetect = 30 * time.Second

func (p *Process) monitor(io garden.ProcessIO) {
	// Be sure that when monitor exists, we have set a finished state
	defer p.setFinished(255)

	n := p.api

	var q nomad.QueryOptions
	q.WaitTime = 10 * time.Second

	var (
		list  []*nomad.AllocationListStub
		meta  *nomad.QueryMeta
		err   error
		id    string
		start = time.Now()
	)

outer:
	for {
		if time.Since(start) > maxAllocDetect {
			fmt.Fprintf(io.Stderr, "Unable to start nomad allocation: waited 30 seconds")
			return
		}

		list, meta, err = n.Jobs().Allocations(p.job, true, &q)
		if err != nil {
			break
		}

		for _, alloc := range list {
			id = alloc.ID
			break outer
		}

		q.WaitIndex = meta.LastIndex
	}

	if id == "" {
		p.logger.Error("nomad-process-allocation-failed", err, lager.Data{"job": p.job})
		fmt.Fprintf(io.Stderr, "Unable to start nomad allocation: %s", err)
		p.setStatus(255)
		return
	}

	p.logger.Info("nomad-allocation-started", lager.Data{"alloc": id})

	q.WaitIndex = meta.LastIndex
	q.WaitTime = 10 * time.Second

	var alloc *nomad.Allocation

	var (
		seenRunning   bool
		seenEventTime int64
		seenNode      bool
	)

	for {
		alloc, meta, err = n.Allocations().Info(id, &q)
		if err != nil {
			p.logger.Error("nomad-process-allocation-info-failed", err)
			fmt.Fprintf(io.Stderr, "Unable to start nomad allocation: %s", err)
			p.setStatus(255)
			return
		}

		p.logger.Debug("nomad-process-alloc-status", lager.Data{
			"alloc":  alloc.ID,
			"status": alloc.ClientStatus,
		})

		if !seenNode && alloc.NodeID != "" {
			seenNode = true
			node, _, err := n.Nodes().Info(alloc.NodeID, nil)
			if err == nil {
				values := []string{
					fmt.Sprintf("id=%s", node.ID),
					fmt.Sprintf("name='%s'", node.Name),
				}

				if ip, ok := node.Attributes["unique.network.ip-address"]; ok {
					values = append(values, fmt.Sprintf("ip=%s", ip))
				}

				if dv, ok := node.Attributes["driver.docker.version"]; ok {
					values = append(values, fmt.Sprintf("docker=%s", dv))
				}

				fmt.Fprintf(io.Stderr, "\x1B[2mnomad: node-info: %s\x1B[0m\n", strings.Join(values, " "))
			} else {
				fmt.Fprintf(io.Stderr, "\x1B[2mnomad: Error reading node info: %s\x1B[0m\n", err)
			}
		}

		for _, ts := range alloc.TaskStates {
			for _, ev := range ts.Events {
				if ev.Time > seenEventTime {
					if ev.Type == "Terminated" && ev.ExitCode == 0 {
						continue
					}

					if len(ev.DisplayMessage) > 0 {
						fmt.Fprintf(io.Stderr, "\x1B[2mnomad: %s\x1B[0m\n", ev.DisplayMessage)
					}
					seenEventTime = ev.Time
				}
			}
		}

		switch alloc.ClientStatus {
		case "running":
			if !seenRunning {
				p.logger.Info("nomad-allocation-running", lager.Data{"alloc": id, "job": alloc.JobID})
				seenRunning = true
			}

		case "complete":
			p.logger.Info("nomad-allocation-finished", lager.Data{"job": p.job, "exit": p.status})

			// Purge completed check jobs because they create a huge amount of clutter in nomad otherwise
			if p.Container.md.Type == db.ContainerTypeCheck {
				p.logger.Debug("nomad-check-job-purged", lager.Data{"job:": p.job})
				n.Jobs().Deregister(p.job, true, nil)
			}
			return
		case "failed":
			p.logger.Error("nomad-allocation-failed", errors.New("allocation failed"), lager.Data{"job": p.job})
			p.setStatus(255)
			return
		}

		q.WaitIndex = meta.LastIndex
	}
}

func (p *Process) setStatus(code int32) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.status = int(code)
}

func (p *Process) setFinished(code int32) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.finished {
		return
	}

	p.finished = true

	if p.status < 0 {
		p.status = int(code)
	}

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

func (p *Process) closeActions() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.actions == nil {
		return
	}

	close(p.actions)

	p.actions = nil
}

func (p *Process) sendAction(act *rpc.Actions) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.actions == nil {
		return
	}

	p.actions <- act
}

func (p *Process) SetTTY(spec garden.TTYSpec) error {
	if spec.WindowSize == nil {
		return nil
	}

	p.sendAction(&rpc.Actions{
		Winsz: &rpc.WindowSize{
			Rows:    int32(spec.WindowSize.Rows),
			Columns: int32(spec.WindowSize.Columns),
		},
	})

	return nil
}

func (p *Process) Signal(sig garden.Signal) error {
	switch sig {
	case garden.SignalKill:
		p.logger.Info("nomad-process-kill", lager.Data{"job": p.job})
		_, _, err := p.api.Jobs().Deregister(p.job, false, &nomad.WriteOptions{})
		return err
	case garden.SignalTerminate:
		p.logger.Info("nomad-process-send-terminate", lager.Data{"job": p.job})
		p.sendAction(&rpc.Actions{
			Signal: 15,
		})
	}

	return nil
}

var _ = garden.Process(&Process{})
