package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/krasin/spectrum-sim/api"
)

var port = flag.Int("port", 2438, "TCP port to listen")

func handle(conn net.Conn) {
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
			panic("send: not implemented")
		case api.Listen:
			panic("listen: not implemented")
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
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Accept: %v", err)
		}
		go handle(conn)
	}
}
