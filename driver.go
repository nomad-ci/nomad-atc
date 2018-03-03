package nomadatc

import (
	"archive/tar"
	"bytes"
	"errors"
	fmt "fmt"
	io "io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	strconv "strconv"
	"sync"
	"sync/atomic"
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/concourse/atc/worker"
	"github.com/golang/snappy"
	lru "github.com/hashicorp/golang-lru"
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/nomad-ci/nomad-atc/assets"
	"github.com/nomad-ci/nomad-atc/rpc"
)

// Driver provides the RPC interface that the drivers running in the builds
// to communicate with.
type Driver struct {
	xid int64

	logger lager.Logger

	serv *grpc.Server

	mu         sync.Mutex
	nextStream int64
	streams    map[int64]*Process

	l       net.Listener
	ldriver net.Listener

	containers *lru.ARCCache

	workDir string
}

// NewDriver creates a new Driver
func NewDriver(logger lager.Logger, workDir string, port int) (*Driver, error) {
	c, err := lru.NewARC(1024)
	if err != nil {
		return nil, err
	}

	d := &Driver{
		logger:     logger.Session("nomad-driver"),
		streams:    make(map[int64]*Process),
		nextStream: 1,
		xid:        time.Now().Unix(),
		containers: c,
		workDir:    workDir,
	}

	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return nil, err
	}

	d.l = l

	ld, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port+1))
	if err != nil {
		return nil, err
	}

	d.ldriver = ld

	d.serv = grpc.NewServer()

	rpc.RegisterTaskServer(d.serv, d)

	return d, nil
}

// XID returns the next unique number
func (d *Driver) XID() int64 {
	return atomic.AddInt64(&d.xid, 1)
}

// Run the driver's components, such as the http server and GRPC server
func (d *Driver) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	close(ready)

	dhs := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.URL.Path == "/_driver.tar.gz" {
				bytes, err := assets.Asset("driver.tar.gz")
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}

				w.Header().Add("Content-Length", strconv.Itoa(len(bytes)))

				w.Write(bytes)
			} else {
				http.NotFound(w, req)
			}
		}),
	}

	errc := make(chan error, 2)

	go func() {
		errc <- dhs.Serve(d.ldriver)
	}()

	go func() {
		errc <- d.serv.Serve(d.l)
	}()

	select {
	case err := <-errc:
		return err
	case <-signals:
		d.l.Close()
		d.ldriver.Close()

		return nil
	}
}

// Register tracks a new Process
func (d *Driver) Register(p *Process) int64 {
	d.mu.Lock()

	sid := d.XID()

	d.streams[sid] = p

	d.mu.Unlock()

	return sid
}

// Deregister removes tracking of a Process
func (d *Driver) Deregister(s int64) {
	d.mu.Lock()

	proc, ok := d.streams[s]
	if ok {
		delete(d.streams, s)

		if proc.cancel != nil {
			proc.cancel()
		}
	}

	d.mu.Unlock()
}

// CancelForBuild finds all the running Processes for the given build and
// closes them
func (d *Driver) CancelForBuild(id int) {
	d.logger.Info("nomad-driver-cancel-for-build", lager.Data{"build": id})
	d.mu.Lock()
	defer d.mu.Unlock()

	var toDelete []int64

	for k, s := range d.streams {
		if s.buildId == id {
			d.logger.Info("nomad-driver-cancel-for-job", lager.Data{"job": s.job})

			if s.cancel != nil {
				s.cancel()
			}

			toDelete = append(toDelete, k)
		}
	}

	for _, k := range toDelete {
		delete(d.streams, k)
	}

	d.logger.Debug("nomad-driver-remove-build-volumes", lager.Data{"build": id})
	os.RemoveAll(filepath.Join(d.workDir, fmt.Sprintf("%d", id)))
}

// FindProcess returns the Process that matches the requested stream id
func (d *Driver) FindProcess(stream int64) (*Process, bool) {
	d.mu.Lock()
	proc, ok := d.streams[int64(stream)]
	d.mu.Unlock()

	return proc, ok
}

// StreamBuildVolume finds the data previously written by SendVolume and returns it.
func (d *Driver) StreamBuildVolume(buildID int, vol, path string) (io.ReadCloser, error) {
	diskpath := filepath.Join(d.workDir, fmt.Sprintf("%d", buildID), vol)
	f, err := os.Open(diskpath)
	if err != nil {
		return nil, err
	}

	if path == "." {
		return f, nil
	}

	defer f.Close()

	sr := snappy.NewReader(f)
	tr := tar.NewReader(sr)

	for {
		hdr, err := tr.Next()
		if err != nil {
			return nil, err
		}

		if hdr.Name == path {
			var buf bytes.Buffer

			tw := tar.NewWriter(&buf)
			tw.WriteHeader(hdr)
			io.Copy(tw, tr)
			tw.Close()

			return ioutil.NopCloser(&buf), nil
		}
	}
}

// SendVolume is called by a per-build driver when it wants to upload a volume.
func (d *Driver) SendVolume(s rpc.Task_SendVolumeServer) error {
	hdr, err := s.Recv()
	if err != nil {
		return err
	}

	proc, ok := d.FindProcess(hdr.Key)
	if !ok {
		return errors.New("unknown process")
	}

	dir := filepath.Join(d.workDir, fmt.Sprintf("%d", proc.buildId))
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(dir, hdr.Name))
	if err != nil {
		return err
	}

	defer f.Close()

	var total int64

	for {
		msg, err := s.Recv()
		if err != nil {
			if err == io.EOF {
				return s.SendAndClose(&rpc.VolumeResponse{
					Size_: total,
				})
			}

			return err
		}

		total += int64(len(msg.Data))

		_, err = f.Write(msg.Data)
		if err != nil {
			return err
		}
	}
}

// ProvideFiles is called via GRPC from a remote driver. This is used to
// allow files to be requested from the remote driver to move volumes
// between containers.
func (d *Driver) ProvideFiles(s rpc.Task_ProvideFilesServer) error {
	md, ok := metadata.FromIncomingContext(s.Context())
	if !ok {
		return errors.New("no stream provided")
	}

	ss, ok := md["stream"]
	if !ok {
		return errors.New("no stream in metadata")
	}

	stream, err := strconv.Atoi(ss[0])
	if err != nil {
		return err
	}

	proc, ok := d.FindProcess(int64(stream))

	if !ok {
		return errors.New("unknown process")
	}

	ctx, cancel := context.WithCancel(s.Context())

	proc.cancel = cancel

	for {
		d.logger.Info("providing-files", lager.Data{"job": proc.job})

		req, err := proc.NextFileRequest(ctx)
		if err != nil {
			break
		}

		start := time.Now()
		d.logger.Info("start-requesting-file", lager.Data{"job": proc.job, "path": req.Path})

		err = s.Send(&rpc.FileRequest{
			Path: req.Path,
		})

		if err != nil {
			return err
		}

		for {
			data, err := s.Recv()
			if err != nil {
				return err
			}

			if len(data.Data) == 0 {
				d.logger.Info("finished-requesting-file", lager.Data{"job": proc.job, "path": req.Path, "elapse": time.Since(start).String()})
				req.Close()
				break
			} else {
				req.Data(data.Data)
			}
		}
	}

	return nil
}

// EmitOutput is called via GRPC by a remote driver. It allows the remote driver
// to send data from streams such as Stdout and Stderr. That data is then fed to
// the process represented by the remote driver.
func (d *Driver) EmitOutput(s rpc.Task_EmitOutputServer) error {
	md, ok := metadata.FromIncomingContext(s.Context())
	if !ok {
		return errors.New("no stream provided")
	}

	ss, ok := md["stream"]
	if !ok {
		return errors.New("no stream in metadata")
	}

	stream, err := strconv.Atoi(ss[0])
	if err != nil {
		return err
	}

	proc, ok := d.FindProcess(int64(stream))

	if !ok {
		return errors.New("unknown process")
	}

	d.logger.Info("nomad-driver-start-output", lager.Data{"job": proc.job})
	defer d.logger.Info("nomad-driver-finish-output", lager.Data{"job": proc.job})

	ctx, cancel := context.WithCancel(s.Context())

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case act := <-proc.actions:
				s.Send(act)
			case <-ctx.Done():
				return
			}
		}
	}()

	if proc.io.Stdin != nil {
		c := make(chan []byte)

		go func() {
			defer close(c)

			for {
				bytes := make([]byte, 1024)
				n, err := proc.io.Stdin.Read(bytes)
				if err != nil {
					return
				}

				select {
				case c <- bytes[:n]:
					// Double check, since we might be done AND have data
					select {
					case <-ctx.Done():
						return
					default:
						// ok, keep going!
					}
				case <-ctx.Done():
					// bye!
					return
				}
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case bytes := <-c:
					if bytes == nil {
						proc.actions <- &rpc.Actions{
							CloseInput: true,
						}
						return
					}

					proc.actions <- &rpc.Actions{
						Input: bytes,
					}
				}
			}
		}()
	}

	for {
		msg, err := s.Recv()
		if err != nil {
			if s, ok := status.FromError(err); ok {
				if s.Code() == codes.Canceled {
					break
				}
			}

			if err != context.Canceled && err != io.EOF {
				d.logger.Error("nomad-driver-output-error", err)
			}

			break
		}

		if len(msg.Data) > 0 {
			switch msg.StreamType {
			case rpc.STDERR:
				if proc.io.Stderr != nil {
					proc.io.Stderr.Write(msg.Data)
				}
			case rpc.STDOUT:
				if proc.io.Stdout != nil {
					proc.io.Stdout.Write(msg.Data)
				}
			case rpc.DRIVER:
				if proc.io.Stderr != nil {
					fmt.Fprintf(proc.io.Stderr, "\x1B[2mnomad: %s\x1B[0m\n", string(msg.Data))
				}
			}
		}

		if msg.Finished {
			d.logger.Info("nomad-driver-process-finished", lager.Data{"job": proc.job, "status": msg.FinishedStatus})
			proc.setFinished(msg.FinishedStatus)
		}
	}

	// get our stdin goroutines to give up the goat
	cancel()

	// and wait for them
	wg.Wait()

	proc.closeActions()

	return nil
}

const chunkSize = 4096

// RequestVolume is called via GRPC by a remote driver. It streams the data
// for the requested volume back.
func (d *Driver) RequestVolume(req *rpc.VolumeRequest, stream rpc.Task_RequestVolumeServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return errors.New("no stream provided")
	}

	ss, ok := md["stream"]
	if !ok {
		return errors.New("no stream in metadata")
	}

	id, err := strconv.Atoi(ss[0])
	if err != nil {
		return err
	}

	proc, ok := d.FindProcess(int64(id))

	if !ok {
		return errors.New("unknown process")
	}

	var found worker.InputSource

	for _, input := range proc.Container.spec.Inputs {
		if input.DestinationPath() == req.Name {
			found = input
			break
		}
	}

	if found == nil {
		return fmt.Errorf("Unable to find mount: %s", req.Name)
	}

	d.logger.Info("start-streaming-volume", lager.Data{"name": req.Name, "job": proc.job})
	defer d.logger.Info("finished-streaming-volume", lager.Data{"name": req.Name, "job": proc.job})

	vs := volStream{proc.job, stream}

	return found.Source().StreamTo(vs)
}

type volStream struct {
	job    string
	stream rpc.Task_RequestVolumeServer
}

func (vs volStream) StreamIn(path string, r io.Reader) error {
	buf := make([]byte, chunkSize)

	for {
		n, err := r.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		err = vs.stream.Send(&rpc.FileData{
			Data: buf[:n],
		})

		if err != nil {
			return err
		}
	}

	return nil
}
