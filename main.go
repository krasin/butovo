package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

var port = flag.Int("port", 2438, "TCP port to listen")

type command int

const (
	send   command = 0
	listen command = 1
)

func read(r io.Reader) (cmd command, ch int, data []byte, err error) {
	panic("read: not implemented")
}

func handle(conn net.Conn) {
	fmt.Printf("Conn: %+v\n", conn)
	defer conn.Close()

	for {
		cmd, _, _, err := read(conn)
		if err != nil {
			log.Printf("Client %v: %v", conn.RemoteAddr(), err)
			return
		}
		switch cmd {
		case send:
			panic("send: not implemented")
		case listen:
			panic("listen: not implemented")
		default:
			log.Printf("Client %v: unknown command %d", conn.RemoteAddr(), cmd)
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
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Accept: %v", err)
		}
		go handle(conn)
	}
}
