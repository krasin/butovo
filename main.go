package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

const (
	minPacketSize = 8
	maxPacketSize = 128
)

var port = flag.Int("port", 2438, "TCP port to listen")

type command int

const (
	send   command = 0
	listen command = 1
)

func read(r io.Reader) (cmd command, ch int, data []byte, err error) {
	var size uint32
	if err = binary.Read(r, binary.LittleEndian, &size); err != nil {
		err = fmt.Errorf("Could not read packet size: %v", err)
		return
	}
	if size > maxPacketSize {
		err = fmt.Errorf("Packet size too large: %d. Max packet size: %d", size, maxPacketSize)
		return
	}
	if size < minPacketSize {
		err = fmt.Errorf("Packet size too small: %d. Min packet size: %d", size, minPacketSize)
		return
	}
	data = make([]byte, size)
	if _, err = io.ReadFull(r, data); err != nil {
		data = nil
		err = fmt.Errorf("Could not read packet body (size: %d): %v", size, err)
		return
	}
	cmd = command(uint32(data[0]) + uint32(data[1])<<8 + uint32(data[2])<<16 + uint32(data[3])<<24)
	ch = int(data[0]) + int(data[1])<<8 + int(data[2])<<16 + int(data[3])<<24
	data = data[8:]
	return
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
