package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/golang/snappy"
	"github.com/kr/pty"
	"github.com/nomad-ci/nomad-atc/config"
	"github.com/nomad-ci/nomad-atc/rpc"
	"github.com/nomad-ci/nomad-atc/tarfs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// So defer's work
func main() {
	realmain()
}

func realmain() {
	log.Printf("driver started")
	path := os.Getenv("NOMAD_META_config")

	var cfg config.TaskConfig

	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	err = json.NewDecoder(f).Decode(&cfg)

	f.Close()

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("config loaded: %#v", &cfg)

	if cfg.Dir != "" {
		log.Printf("running in: %s", cfg.Dir)
		err = os.MkdirAll(cfg.Dir, 755)
		if err != nil {
			log.Fatal(err)
		}

		err = os.Chdir(cfg.Dir)
		if err != nil {
			log.Fatal(err)
		}
	}

	conn, err := grpc.Dial(cfg.Host, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	tc := rpc.NewTaskClient(conn)

	md := metadata.Pairs("stream", strconv.Itoa(int(cfg.Stream)))

	ctx := metadata.NewOutgoingContext(context.Background(), md)

	emit, err := tc.EmitOutput(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if len(cfg.Inputs) > 0 {
		log.Printf("inputs to consume: %s", cfg.Inputs)

		var wg sync.WaitGroup

		wg.Add(len(cfg.Inputs))

		total := new(int64)
		start := time.Now()

		for _, vol := range cfg.Inputs {
			go func(vol *config.Volume) {
				defer wg.Done()

				fd, err := tc.RequestVolume(ctx, &rpc.VolumeRequest{
					Name: vol.Path,
				})

				fdtr := &filedataToReader{files: fd}

				sr := snappy.NewReader(fdtr)

				err = tarfs.Extract(sr, vol.Path)
				if err != nil {
					log.Fatal(err)
				}

				atomic.AddInt64(total, int64(fdtr.totalBytes))
			}(vol)
		}

		wg.Wait()

		dur := time.Since(start)

		bps := float64(*total) / dur.Seconds()

		unit := "B"

		switch {
		case bps > 1024*1024:
			bps = bps / (1024 * 1024)
			unit = "MB"
		case bps > 1024:
			bps = bps / 1024
			unit = "KB"
		default:
		}

		data := fmt.Sprintf("driver: read volumes at %.2f %s/s (%d in %s)", bps, unit, *total, dur)

		emit.Send(&rpc.OutputData{
			Stream:     cfg.Stream,
			StreamType: rpc.DRIVER,
			Data:       []byte(data),
		})

		fmt.Println(data)
	}

	for _, vol := range cfg.Outputs {
		log.Printf("creating output dir: %s", vol.Path)
		os.MkdirAll(vol.Path, 0755)
	}

	cmd := exec.Command(cfg.Path, cfg.Args...)
	cmd.Env = append(cfg.Env, os.Environ()...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	var (
		stdout    io.Reader
		tty, lpty *os.File
	)

	if cfg.UseTTY {
		lpty, tty, err = pty.Open()
		if err != nil {
			log.Fatal(err)
		}

		defer lpty.Close()

		cmd.Stdout = tty

		if winsz := cfg.WindowSize; winsz != nil {
			pty.Setsize(lpty, &pty.Winsize{
				Rows: uint16(winsz.Rows),
				Cols: uint16(winsz.Columns),
			})
		}

		stdout = lpty
	} else {
		stdout, err = cmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}
	}

	var (
		buffers sync.Pool
		output  = make(chan *rpc.OutputData)
		done    = make(chan bool, 2)
	)

	buffers.New = func() interface{} {
		return make([]byte, 1024)
	}

	go func() {
		defer func() {
			done <- true
		}()

		for {
			bytes := buffers.Get().([]byte)
			n, rerr := stdout.Read(bytes)

			if n > 0 {
				os.Stdout.Write(bytes[:n])

				output <- &rpc.OutputData{
					Stream:     cfg.Stream,
					StreamType: rpc.STDOUT,
					Data:       bytes[:n],
				}
			}

			if rerr != nil {
				return
			}
		}
	}()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer func() {
			done <- true
		}()

		for {
			bytes := buffers.Get().([]byte)
			n, rerr := stderr.Read(bytes)

			if n > 0 {
				os.Stderr.Write(bytes[:n])

				output <- &rpc.OutputData{
					Stream:     cfg.Stream,
					StreamType: rpc.STDERR,
					Data:       bytes[:n],
				}
			}

			if rerr != nil {
				return
			}
		}
	}()

	err = cmd.Start()
	if err != nil {
		log.Printf("error starting program: %s", err)

		log.Printf("Unable to start requested program, sending error state to concourse")

		emit.Send(&rpc.OutputData{
			Stream:         cfg.Stream,
			StreamType:     rpc.STDERR,
			Data:           []byte(fmt.Sprintf("Unable to run %s: %s\n", cfg.Path, err)),
			Finished:       true,
			FinishedStatus: 255,
		})

		emit.CloseSend()

		for {
			_, err := emit.Recv()
			if err != nil {
				break
			}
		}

		conn.Close()

		return
	}

	// So that only the created process owns it's tty
	if tty != nil {
		tty.Close()
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			msg, err := emit.Recv()
			if err != nil {
				return
			}

			if msg == nil {
				return
			}

			if msg.Signal > 0 {
				var sig os.Signal
				switch msg.Signal {
				case 1:
					sig = syscall.SIGHUP
				case 2:
					sig = syscall.SIGINT
				case 3:
					sig = syscall.SIGQUIT
				case 9:
					sig = syscall.SIGKILL
				case 15:
					sig = syscall.SIGTERM
				case 30:
					sig = syscall.SIGUSR1
				case 31:
					sig = syscall.SIGUSR2
				default:
					log.Printf("unknown signal, using sigterm: %d", msg.Signal)
					sig = syscall.SIGTERM
				}

				log.Printf("sending process signal: %d", sig)
				cmd.Process.Signal(sig)
			}

			if lpty != nil {
				if winsz := msg.Winsz; winsz != nil {
					pty.Setsize(lpty, &pty.Winsize{
						Rows: uint16(winsz.Rows),
						Cols: uint16(winsz.Columns),
					})
				}
			}

			for len(msg.Input) > 0 {
				n, err := stdin.Write(msg.Input)
				if err != nil {
					log.Printf("error writing to stdin: %s", err)
					break
				}

				msg.Input = msg.Input[n:]
			}

			if msg.CloseInput {
				log.Printf("closing stdin")
				stdin.Close()
			}
		}
	}()

	log.Printf("waiting on output...")
	left := 2

	for left > 0 {
		select {
		case data := <-output:
			err = emit.Send(data)
			if err != nil {
				log.Fatal(err)
			}
			// The :cap() slice is to re-expand the byte slice to
			// it's full width now that we've used the data that was
			// set.
			buffers.Put(data.Data[:cap(data.Data)])
		case <-done:
			left--
		}
	}

	log.Printf("waiting on process...")
	err = cmd.Wait()

	if err != nil {
		log.Printf("sending finished with status=1")
		err = emit.Send(&rpc.OutputData{
			Finished:       true,
			FinishedStatus: 1,
		})

		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Printf("sending volumes")

		for _, vol := range cfg.Outputs {
			sender, err := tc.SendVolume(ctx)
			if err != nil {
				log.Fatal(err)
			}

			err = sender.Send(&rpc.VolumeData{
				Key:  cfg.Stream,
				Name: vol.Handle,
			})

			if err != nil {
				log.Fatal(err)
			}

			var wdt writerToData
			wdt.files = sender

			tarDir := vol.Path
			tarPath := "."

			sw := snappy.NewBufferedWriter(wdt)

			err = tarfs.Compress(sw, tarDir, tarPath)
			if err != nil {
				log.Printf("unable to tar  %s: %s", tarDir, err)
				sender.CloseAndRecv()
				continue
			}

			sw.Close()

			resp, err := sender.CloseAndRecv()
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("sent '%s': %d bytes reported", vol.Handle, resp.Size_)
		}

		log.Printf("sending finished with status=0")
		err = emit.Send(&rpc.OutputData{
			Finished:       true,
			FinishedStatus: 0,
		})

		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("process done: %s", err)

	emit.CloseSend()
	wg.Wait()

	conn.Close()

	log.Printf("exitting: %v", conn.GetState())
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}

type writerToData struct {
	files rpc.Task_SendVolumeClient
}

const chunkSize = 1024 * 32

func (w writerToData) Write(data []byte) (int, error) {
	n := len(data)

	for len(data) > 0 {
		var buf []byte

		if len(data) > chunkSize {
			buf = data[:chunkSize]
			data = data[chunkSize:]
		} else {
			buf = data
			data = nil
		}

		err := w.files.Send(&rpc.VolumeData{
			Data: buf,
		})

		if err != nil {
			return 0, nil
		}
	}

	return n, nil
}

type filedataToReader struct {
	files rpc.Task_RequestVolumeClient
	cont  *rpc.FileData

	totalBytes int
}

const showEveryBytes = 1024 * 1024

func (f *filedataToReader) Read(b []byte) (int, error) {
	if f.cont != nil {
		if len(b) < len(f.cont.Data) {
			copy(b, f.cont.Data[:len(b)])
			f.cont.Data = f.cont.Data[len(b):]
			return len(b), nil
		}

		data := f.cont
		f.cont = nil
		copy(b, data.Data)
		return len(data.Data), nil
	}

	data, err := f.files.Recv()
	if err != nil {
		return 0, err
	}

	f.totalBytes += len(data.Data)

	if len(data.Data) == 0 {
		return 0, io.EOF
	}

	if len(data.Data) > len(b) {
		copy(b, data.Data[:len(b)])
		data.Data = data.Data[len(b):]
		f.cont = data
		return len(b), nil
	}

	copy(b, data.Data)
	return len(data.Data), nil
}
