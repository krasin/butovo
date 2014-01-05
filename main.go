package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/krasin/spectrum-sim/api"
)

var port = flag.Int("port", 2438, "TCP port to listen")

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

func runChannel(id int, ch <-chan chanReq, errChan chan<- error) {
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

func newChannel(id int, errChan chan<- error) chan<- chanReq {
	ch := make(chan chanReq)
	go runChannel(id, ch, errChan)
	return ch
}

type server struct {
	m       sync.Mutex
	errChan chan<- error
	chans   map[int]chan<- chanReq
}

func handleErrors(errChan <-chan error) {
	for err := range errChan {
		log.Print(err)
	}
}

func newServer() *server {
	errChan := make(chan error)

	go handleErrors(errChan)
	return &server{
		chans:   make(map[int]chan<- chanReq),
		errChan: errChan,
	}
}

func (s *server) getChannel(ch int) chan<- chanReq {
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

func (s *server) handle(conn net.Conn) {
	fmt.Printf("Conn: %+v\n", conn)
	defer conn.Close()

	recvCh := make(chan []byte, 1)
	closeCh := make(chan bool)
	go runSender(conn, recvCh, closeCh, s.errChan)
	defer close(closeCh)

	curCh := -1
	key := fmt.Sprintf("key-%d", time.Now().UnixNano())

	forget := func() {
		if curCh < 0 {
			return
		}

		s.getChannel(curCh) <- chanReq{
			cmd: chanForget,
			key: key,
		}
		curCh = -1
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

func main() {
	flag.Parse()

	if port == nil {
		flag.PrintDefaults()
		os.Exit(1)
	}
	laddr := fmt.Sprintf(":%d", *port)
	ln, err := net.Listen("tcp", laddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Listen(tcp, %q): %v", laddr, err)
		os.Exit(1)
	}
	fmt.Printf("Serving on port %d...\n", *port)
	s := newServer()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Accept: %v", err)
		}
		go s.handle(conn)
	}
}
