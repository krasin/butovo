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
	w    io.Writer
}

func runChannel(name string, ch <-chan chanReq) {
	w := make(map[string]io.Writer)

	for req := range ch {
		log.Printf("%q: %+v", name, req)
		switch req.cmd {
		case chanSend:
			log.Printf("chanSend not implemented")
		case chanListen:
			w[req.key] = req.w
		case chanForget:
			delete(w, req.key)
		}
	}
}

func newChannel(name string) chan<- chanReq {
	ch := make(chan chanReq)
	go runChannel(name, ch)
	return ch
}

type server struct {
	m     sync.Mutex
	chans map[int]chan<- chanReq
}

func newServer() *server {
	return &server{chans: make(map[int]chan<- chanReq)}
}

func (s *server) getChannel(ch int) chan<- chanReq {
	s.m.Lock()
	defer s.m.Unlock()
	if s.chans[ch] == nil {
		s.chans[ch] = newChannel(fmt.Sprintf("channel %d", ch))
	}
	return s.chans[ch]
}

func (s *server) handle(conn net.Conn) {
	fmt.Printf("Conn: %+v\n", conn)
	defer conn.Close()

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
		cmd, err := api.Read(conn)
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
			}
		case api.Listen:
			forget()
			ch <- chanReq{
				cmd: chanListen,
				key: key,
				w:   conn,
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
