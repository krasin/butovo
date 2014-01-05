// Package sim provides a server that accepts Listen and Send request to numbered channels
// and multiplexes the sent data to the listeners. It's supposed to be useful for Bluetooth Low Energy
// stacks testing, because it provides a packet level abstraction of the spectrum.
// The wire format is simple enough to be implemented in plain C.
// It's documented in github.com/krasin/butovo/tools/spectrum-sim/api package.
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
			resp := &api.Response{
				Channel:   id,
				Timestamp: req.ts,
				Data:      req.data,
			}
			data, err := api.WriteResponse(resp)
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
		req, err := api.ReadRequest(conn)
		if err != nil {
			log.Printf("Client %v: %v", conn.RemoteAddr(), err)
			return
		}

		switch req.Type {
		case api.Send:
			forget()
			ch := s.getChannel(req.Channel, false)
			if ch == nil {
				// no listeners
				break
			}
			ch <- chanReq{
				cmd:  chanSend,
				data: req.Data,
				ts:   time.Now().UTC(),
			}
		case api.Listen:
			forget()
			ch := s.getChannel(req.Channel, true)
			ch <- chanReq{
				cmd: chanListen,
				key: key,
				to:  recvCh,
			}
			curCh = req.Channel
		default:
			log.Printf("Client %v: unknown command %d", conn.RemoteAddr(), req.Data)
			return
		}
	}
}
