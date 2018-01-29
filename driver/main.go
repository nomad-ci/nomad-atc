package main

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/concourse/go-archive/tarfs"
	nomadatc "github.com/nomad-ci/nomad-atc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func main() {
	log.Printf("driver started")
	path := os.Getenv("NOMAD_META_config")

	var cfg nomadatc.TaskConfig

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

	tc := nomadatc.NewTaskClient(conn)

	md := metadata.Pairs("stream", strconv.Itoa(int(cfg.Stream)))

	ctx := metadata.NewOutgoingContext(context.Background(), md)

	if len(cfg.Inputs) > 0 {
		log.Printf("inputs to consume: %s", cfg.Inputs)

		for _, name := range cfg.Inputs {
			fd, err := tc.RequestVolume(ctx, &nomadatc.VolumeRequest{
				Name: name,
			})

			gzr, err := gzip.NewReader(filedataToReader{files: fd})
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("extracting to %s", name)
			err = tarfs.Extract(gzr, name)
			gzr.Close()
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	emit, err := tc.EmitOutput(ctx)
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command(cfg.Path, cfg.Args...)
	cmd.Env = cfg.Env

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	var (
		buffers sync.Pool
		output  = make(chan *nomadatc.OutputData)
		done    = make(chan bool, 2)
	)

	buffers.New = func() interface{} {
		return make([]byte, 1024)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
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

				output <- &nomadatc.OutputData{
					Stream: cfg.Stream,
					Data:   bytes[:n],
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

				output <- &nomadatc.OutputData{
					Stream: cfg.Stream,
					Stderr: true,
					Data:   bytes[:n],
				}
			}

			if rerr != nil {
				return
			}
		}
	}()

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
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

				cmd.Process.Signal(sig)
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
			buffers.Put(data.Data)
		case <-done:
			left--
		}
	}

	log.Printf("waiting on process...")
	err = cmd.Wait()

	if err != nil {
		log.Printf("sending finished with status=1")
		err = emit.Send(&nomadatc.OutputData{
			Finished:       true,
			FinishedStatus: 1,
		})

		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Printf("sending finished with status=0")
		err = emit.Send(&nomadatc.OutputData{
			Finished:       true,
			FinishedStatus: 0,
		})

		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("process done: %s", err)

	if cfg.WaitForVolumes {
		ctx, _ = context.WithTimeout(ctx, 10*time.Minute)

		files, err := tc.ProvideFiles(ctx)
		if err != nil {
			log.Fatal(err)
		}

		for {
			req, err := files.Recv()
			if err != nil {
				log.Printf("files recv error: %s", err)
				break
			}

			log.Printf("fetching requested file: %s", req.Path)

			var wdt writerToData
			wdt.files = files

			var out io.Writer = wdt

			src := req.Path

			fileInfo, err := os.Stat(src)
			if err != nil {
				log.Fatal(err)
			}

			var tarDir, tarPath string
			var gzw *gzip.Writer

			if fileInfo.IsDir() {
				tarDir = src
				tarPath = "."

				gzw = gzip.NewWriter(out)
				out = gzw
			} else {
				tarDir = filepath.Dir(src)
				tarPath = filepath.Base(src)
			}

			err = tarfs.Compress(out, tarDir, tarPath)
			if err != nil {
				log.Fatal(err)
			}

			if gzw != nil {
				gzw.Close()
			}

			err = files.Send(&nomadatc.FileData{})
			if err != nil {
				log.Fatal(err)
			}
		}
	}

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
	files nomadatc.Task_ProvideFilesClient
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

		err := w.files.Send(&nomadatc.FileData{
			Data: buf,
		})

		if err != nil {
			return 0, nil
		}
	}

	return n, nil
}

type filedataToReader struct {
	files nomadatc.Task_RequestVolumeClient
	cont  *nomadatc.FileData
}

func (f filedataToReader) Read(b []byte) (int, error) {
	if f.cont != nil {
		if len(b) < len(f.cont.Data) {
			copy(b, f.cont.Data[:len(b)])
			f.cont.Data = f.cont.Data[len(b):]
			return len(b), nil
		} else {
			data := f.cont
			f.cont = nil
			copy(b, data.Data)
			return len(data.Data), nil
		}
	}

	data, err := f.files.Recv()
	if err != nil {
		return 0, err
	}

	if len(data.Data) == 0 {
		return 0, io.EOF
	}

	if len(data.Data) > len(b) {
		copy(b, data.Data[:len(b)])
		data.Data = data.Data[len(b):]
		f.cont = data
		return len(b), nil
	} else {
		copy(b, data.Data)
		return len(data.Data), nil
	}
}
