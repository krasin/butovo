package sim

import (
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"sync"
	"time"

	"github.com/krasin/butovo/tools/spectrum-sim/api"
)

type chanCmd int

const (
	chanSend chanCmd = iota
	chanListen
	chanForget
)

type chanReq struct {
	cmd  chanCmd
	data []byte
	key  string
	to   chan<- []byte
	w    io.Writer
	ts   time.Time
}

func runChannel(id uint32, ch <-chan chanReq, errChan chan<- error) {
	lis := make(map[string]chan<- []byte)

	for req := range ch {
		log.Printf("Channel %d: %+v", id, req)
		switch req.cmd {
		case chanSend:
			data, err := api.WriteResponse(uint32(id), req.ts, req.data)
			if err != nil {
				errChan <- err
				continue
			}
			log.Printf("to send: %v", data)
			for _, to := range lis {
				// It's better to drop some packet, than block on a slow client.
				select {
				case to <- data:
				default:
				}
			}
		case chanListen:
			lis[req.key] = req.to
		case chanForget:
			delete(lis, req.key)
		}
	}
}

func newChannel(id uint32, errChan chan<- error) chan<- chanReq {
	ch := make(chan chanReq)
	go runChannel(id, ch, errChan)
	return ch
}

type Server struct {
	m       sync.Mutex
	errChan chan<- error
	chans   map[uint32]chan<- chanReq
}

func NewServer(errChan chan<- error) *Server {
	return &Server{
		chans:   make(map[uint32]chan<- chanReq),
		errChan: errChan,
	}
}

func (s *Server) getChannel(ch uint32) chan<- chanReq {
	s.m.Lock()
	defer s.m.Unlock()
	if s.chans[ch] == nil {
		s.chans[ch] = newChannel(ch, s.errChan)
	}
	return s.chans[ch]
}

func runSender(w io.Writer, recvCh <-chan []byte, closeCh <-chan bool, errChan chan<- error) {
	for {
		select {
		case data := <-recvCh:
			_, err := w.Write(data)
			if err != nil {
				errChan <- err
			}
		case <-closeCh:
			return
		}
	}
}

func (s *Server) Handle(conn net.Conn) {
	log.Printf("Conn: %+v\n", conn)
	defer conn.Close()

	recvCh := make(chan []byte, 1)
	closeCh := make(chan bool)
	go runSender(conn, recvCh, closeCh, s.errChan)
	defer close(closeCh)

	var curCh uint32 = math.MaxUint32
	key := fmt.Sprintf("key-%d", time.Now().UnixNano())

	forget := func() {
		if curCh < 0 {
			return
		}

		s.getChannel(curCh) <- chanReq{
			cmd: chanForget,
			key: key,
		}
		curCh = math.MaxUint32
	}
	defer forget()

	for {
		cmd, err := api.ReadRequest(conn)
		if err != nil {
			log.Printf("Client %v: %v", conn.RemoteAddr(), err)
			return
		}
		ch := s.getChannel(cmd.Channel)

		switch cmd.Cmd {
		case api.Send:
			forget()
			ch <- chanReq{
				cmd:  chanSend,
				data: cmd.Data,
				ts:   time.Now().UTC(),
			}
		case api.Listen:
			forget()
			ch <- chanReq{
				cmd: chanListen,
				key: key,
				to:  recvCh,
			}
			curCh = cmd.Channel
		default:
			log.Printf("Client %v: unknown command %d", conn.RemoteAddr(), cmd.Cmd)
			return
		}
	}
}
