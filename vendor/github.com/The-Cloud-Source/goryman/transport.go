package goryman

import (
	"bufio"
	"encoding/binary"
	"io"
	"io/ioutil"
	"log"
	"net"

	pb "github.com/golang/protobuf/proto"
	"github.com/The-Cloud-Source/goryman/proto"
)

type Transport struct {
	conn net.Conn
	wb   *bufio.Writer
	w    chan *proto.Msg
	err  chan error
}

// NewTcpTransport - Factory
func NewTcpTransport(conn net.Conn) *Transport {
	t := &Transport{
		conn: conn,
		wb:   bufio.NewWriter(conn),
		w:    make(chan *proto.Msg),
		err:  make(chan error, 1), // fatal error
	}
	go t.writeLoop()
	go t.readLoop()
	return t
}

// Send, queues a message to send a the Riemann server.
func (t *Transport) Send(message *proto.Msg) error {
	select {
	case t.w <- message:
		return nil
	case err := <-t.err:
		log.Print(err)
		return err
	}
}

// Close implements the io.Closer interface.
func (t *Transport) Close() error {
	close(t.w)
	return t.conn.Close()
}

// writeLoop writes incoming messages over TCP.
func (t *Transport) writeLoop() {
	for message := range t.w {
		if err := t.write(message); err != nil {
			message.Events = message.Events[:0]
			log.Print(err)
			t.err <- err
			return
		}
		message.Events = message.Events[:0]
	}
}

// write writes the given message. If a transport error is encountered,
// the error is returned and the client is considered dead.
func (t *Transport) write(message *proto.Msg) error {
	data, err := pb.Marshal(message)
	if err != nil {
		log.Panicf("goryman: marshal error: %v", err)
	}
	if err = binary.Write(t.wb, binary.BigEndian, uint32(len(data))); err != nil {
		log.Print(err)
		return err
	}
	if _, err = t.wb.Write(data); err != nil {
		log.Print(err)
		return err
	}
	return nil
}

// readLoop receives server responses.
func (t *Transport) readLoop() {
	var header uint32
	for {
		if err := binary.Read(t.conn, binary.BigEndian, &header); err != nil {
			return
		}
		_, err := io.CopyN(ioutil.Discard, t.conn, int64(header))
		if err != nil {
			return
		}
	}
}
