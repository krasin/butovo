package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	fd := make([]int, 2)
	if err := syscall.Pipe(fd); err != nil {
		log.Fatal("Socketpair:", err)
	}
	defer syscall.Close(fd[0])
	defer syscall.Close(fd[1])
	name := fmt.Sprintf("/proc/%d/fd/%d", os.Getpid(), fd[1])
	fmt.Println(name)

	cmd := exec.Command("ubertooth-btle", "-f", "-d", name)
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	f := os.NewFile(uintptr(fd[0]), "server")
	buf := make([]byte, 1024)
	for {
		n, err := f.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%q\n", buf[:n])
	}
}
