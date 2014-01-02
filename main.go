package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/krasin/spectrum-sim/api"
)

var port = flag.Int("port", 2438, "TCP port to listen")

type channel struct {
}

func (ch *channel) Send(data []byte) error {
	panic("channel.Send not implemented")
}

func (ch *channel) Listen() error {
	panic("channel.Listen not implemented")
}

func newChannel() *channel {
	return &channel{}
}

type server struct {
	m     sync.Mutex
	chans map[int]*channel
}

func newServer() *server {
	return &server{chans: make(map[int]*channel)}
}

func (s *server) getChannel(ch int) *channel {
	s.m.Lock()
	defer s.m.Unlock()
	if s.chans[ch] == nil {
		s.chans[ch] = newChannel()
	}
	return s.chans[ch]
}

func (s *server) handle(conn net.Conn) {
	fmt.Printf("Conn: %+v\n", conn)
	defer conn.Close()

	for {
		cmd, err := api.Read(conn)
		if err != nil {
			log.Printf("Client %v: %v", conn.RemoteAddr(), err)
			return
		}
		switch cmd.Cmd {
		case api.Send:
			if err = s.getChannel(cmd.Channel).Send(cmd.Data); err != nil {
				log.Printf("Send failed: %v", err)
				return
			}
		case api.Listen:
			if err = s.getChannel(cmd.Channel).Listen(); err != nil {
				log.Printf("Listen failed: %v", err)
				return
			}
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
