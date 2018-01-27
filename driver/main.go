package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

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

	conn, err := grpc.Dial(cfg.Host, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	tc := nomadatc.NewTaskClient(conn)

	ctx := context.Background()

	md := metadata.Pairs("stream", strconv.Itoa(int(cfg.Stream)))

	ctx = metadata.NewOutgoingContext(ctx, md)

	emit, err := tc.EmitOutput(ctx)
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command(cfg.Path, cfg.Args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		bytes := make([]byte, 1024)

		for {
			n, rerr := stdout.Read(bytes)

			if n > 0 {
				os.Stdout.Write(bytes[:n])

				err = emit.Send(&nomadatc.OutputData{
					Stream: cfg.Stream,
					Data:   bytes[:n],
				})

				if err != nil {
					log.Printf("stdout send err: %s", err)
					return
				}
			}

			if rerr != nil {
				log.Printf("stdout err: %s", rerr)
				emit.Send(&nomadatc.OutputData{
					Stream: cfg.Stream,
				})
				return
			}
		}
	}()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		bytes := make([]byte, 1024)

		for {
			n, rerr := stderr.Read(bytes)

			if n > 0 {
				os.Stderr.Write(bytes[:n])

				log.Printf("sending stderr data: %v", string(bytes[:n]))

				err = emit.Send(&nomadatc.OutputData{
					Stream: cfg.Stream,
					Stderr: true,
					Data:   bytes[:n],
				})

				if err != nil {
					log.Printf("stderr send err: %s", err)
					return
				}
			}

			if rerr != nil {
				log.Printf("stderr err: %s", rerr)
				emit.Send(&nomadatc.OutputData{
					Stream: cfg.Stream,
				})
				return
			}
		}
	}()

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
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
	wg.Wait()

	log.Printf("waiting on process...")
	err = cmd.Wait()

	if err != nil {
		err = emit.Send(&nomadatc.OutputData{
			Finished:       true,
			FinishedStatus: 1,
		})

		if err != nil {
			log.Fatal(err)
		}
	} else {
		err = emit.Send(&nomadatc.OutputData{
			Finished:       true,
			FinishedStatus: 0,
		})

		if err != nil {
			log.Fatal(err)
		}
	}

	err = emit.CloseSend()
	if err != nil {
		log.Printf("close and send failed: %s", err)
	}

	log.Printf("process done: %s", err)

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

		f, err := os.Open(req.Path)
		if err != nil {
			log.Printf("Unable to open requested file: %s", err)
			files.Send(&nomadatc.FileData{})
			continue
		}

		var tarBuf bytes.Buffer

		tw := tar.NewWriter(&tarBuf)

		stat, err := f.Stat()
		if err != nil {
			log.Printf("Unable to stat requested file: %s", err)
			files.Send(&nomadatc.FileData{})
			continue
		}

		hdr, err := tar.FileInfoHeader(stat, "")
		if err != nil {
			log.Printf("Unable to make tar header: %s", err)
			files.Send(&nomadatc.FileData{})
			continue
		}

		err = tw.WriteHeader(hdr)
		if err != nil {
			log.Printf("Unable to write tar header: %s", err)
			files.Send(&nomadatc.FileData{})
			continue
		}

		_, err = io.Copy(tw, f)
		if err != nil {
			log.Printf("Unable to copy file to tar: %s", err)
			files.Send(&nomadatc.FileData{})
			continue
		}

		err = tw.Close()
		if err != nil {
			log.Printf("unable to close tar writer: %s", err)
			files.Send(&nomadatc.FileData{})
			continue
		}

		log.Printf("sending file as tar...")

		data := tarBuf.Bytes()

		for len(data) > 0 {
			var buf []byte

			if len(data) > 4096 {
				buf = data[:4096]
				data = data[:4096]
			} else {
				buf = data
				data = nil
			}

			err = files.Send(&nomadatc.FileData{
				Data: buf,
			})

			if err != nil {
				log.Fatal(err)
			}
		}
	}

	err = files.Send(&nomadatc.FileData{})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("exitting.")
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}

func writeTar(source string, output io.Writer) error {
	gzw := gzip.NewWriter(output)
	defer gzw.Close()

	tarWriter := tar.NewWriter(gzw)
	defer tarWriter.Close()

	base := source

	if !strings.HasSuffix(base, "/") {
		base = base + "/"
	}

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking to %s: %v", path, err)
		}

		if path == source {
			return nil
		}

		header, err := tar.FileInfoHeader(info, path)
		if err != nil {
			return fmt.Errorf("%s: making header: %v", path, err)
		}

		header.Name = strings.TrimPrefix(path, base)

		if info.IsDir() && !strings.HasSuffix(header.Name, "/") {
			header.Name += "/"
		}

		err = tarWriter.WriteHeader(header)
		if err != nil {
			return fmt.Errorf("%s: writing header: %v", path, err)
		}

		if info.IsDir() {
			return nil
		}

		if header.Typeflag == tar.TypeReg {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("%s: open: %v", path, err)
			}
			defer file.Close()

			_, err = io.CopyN(tarWriter, file, info.Size())
			if err != nil && err != io.EOF {
				return fmt.Errorf("%s: copying contents: %v", path, err)
			}
		}
		return nil
	})
}
