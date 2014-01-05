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
	log.Printf("runChannel(%d): quit", id)
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
	cnt     map[uint32]int
}

func NewServer(errChan chan<- error) *Server {
	return &Server{
		chans:   make(map[uint32]chan<- chanReq),
		cnt:     make(map[uint32]int),
		errChan: errChan,
	}
}

func (s *Server) getChannel(ch uint32, mayCreate bool) chan<- chanReq {
	s.m.Lock()
	defer s.m.Unlock()
	if s.chans[ch] == nil {
		if !mayCreate {
			return nil
		}
		s.chans[ch] = newChannel(ch, s.errChan)
	}
	s.cnt[ch]++
	return s.chans[ch]
}

func (s *Server) releaseChannel(ch uint32, key string) {
	s.m.Lock()
	defer s.m.Unlock()
	s.chans[ch] <- chanReq{
		cmd: chanForget,
		key: key,
	}
	s.cnt[ch]--
	if s.cnt[ch] == 0 {
		log.Printf("Channel %d does not have any listeners anymore", ch)
		close(s.chans[ch])
		delete(s.chans, ch)
	}
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
	log.Printf("Conn: %+v", conn)
	defer conn.Close()

	recvCh := make(chan []byte, 1)
	closeCh := make(chan bool)
	go runSender(conn, recvCh, closeCh, s.errChan)
	defer close(closeCh)

	var curCh uint32 = math.MaxUint32
	key := fmt.Sprintf("key-%d", time.Now().UnixNano())

	forget := func() {
		if curCh == math.MaxUint32 {
			return
		}
		s.releaseChannel(curCh, key)
		curCh = math.MaxUint32
	}
	defer forget()

	for {
		cmd, err := api.ReadRequest(conn)
		if err != nil {
			log.Printf("Client %v: %v", conn.RemoteAddr(), err)
			return
		}

		switch cmd.Cmd {
		case api.Send:
			forget()
			ch := s.getChannel(cmd.Channel, false)
			if ch == nil {
				// no listeners
				break
			}
			ch <- chanReq{
				cmd:  chanSend,
				data: cmd.Data,
				ts:   time.Now().UTC(),
			}
		case api.Listen:
			forget()
			ch := s.getChannel(cmd.Channel, true)
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
