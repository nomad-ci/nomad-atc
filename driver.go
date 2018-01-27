package nomadatc

import (
	"errors"
	"net"
	"os"
	strconv "strconv"
	"sync"

	"github.com/davecgh/go-spew/spew"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Driver struct {
	root string
	serv *grpc.Server

	mu         sync.Mutex
	nextStream int64
	streams    map[int64]*Process

	l net.Listener
}

func NewDriver(root string) (*Driver, error) {
	d := &Driver{
		root:       root,
		streams:    make(map[int64]*Process),
		nextStream: 1,
	}

	l, err := net.Listen("tcp", "0.0.0.0:12101")
	if err != nil {
		return nil, err
	}

	d.l = l
	d.serv = grpc.NewServer()

	RegisterTaskServer(d.serv, d)

	return d, nil
}

func (d *Driver) Run(signal <-chan os.Signal, ready chan<- struct{}) error {
	close(ready)

	return d.serv.Serve(d.l)
}

func (d *Driver) Register(p *Process) int64 {
	d.mu.Lock()

	sid := d.nextStream
	d.nextStream++

	d.streams[sid] = p

	d.mu.Unlock()

	return sid
}

func (d *Driver) Deregister(s int64) {
	d.mu.Lock()

	delete(d.streams, s)

	d.mu.Unlock()
}

func (d *Driver) FindProcess(stream int64) (*Process, bool) {
	d.mu.Lock()
	proc, ok := d.streams[int64(stream)]
	d.mu.Unlock()

	return proc, ok
}

func (d *Driver) ProvideFiles(s Task_ProvideFilesServer) error {
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

	for {
		req, err := proc.NextFileRequest()
		if err != nil {
			break
		}

		err = s.Send(&FileRequest{
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
				req.Close()
			} else {
				req.Data(data.Data)
			}
		}
	}

	return nil
}

func (d *Driver) EmitOutput(s Task_EmitOutputServer) error {
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

	if proc.io.Stdin != nil {
		go func() {
			for {
				bytes := make([]byte, 1024)
				n, err := proc.io.Stdin.Read(bytes)
				if err != nil {
					s.Send(&Actions{
						CloseInput: true,
					})

					return
				}

				err = s.Send(&Actions{
					Input: bytes[:n],
				})

				if err != nil {
					return
				}
			}
		}()
	}

	for {
		msg, err := s.Recv()
		if err != nil {
			return err
		}

		spew.Dump(msg)
		spew.Dump(msg.Data)

		if msg.Finished {
			proc.setFinished(msg.FinishedStatus)
			continue
		}

		if len(msg.Data) > 0 {
			if msg.Stderr {
				if proc.io.Stderr != nil {
					proc.io.Stderr.Write(msg.Data)
				}
			} else {
				if proc.io.Stdout != nil {
					proc.io.Stdout.Write(msg.Data)
				}
			}
		}
	}
}
