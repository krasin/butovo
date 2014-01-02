package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

var port = flag.Int("port", 2438, "TCP port to listen")

func handle(conn net.Conn) {
	fmt.Printf("Conn: %+v\n", conn)
	conn.Close()
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
