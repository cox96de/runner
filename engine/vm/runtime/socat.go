package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/cox96de/runner/log"
)

func socat(a, b string) {
	partA, err := open(a)
	if err != nil {
		log.Fatal(err)
	}
	partB, err := open(b)
	if err != nil {
		log.Fatal(err)
	}
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, err := io.Copy(partA, partB)
		if err != nil {
			log.Fatal(err)
		}
		partA.Close()
	}()
	go func() {
		defer wg.Done()
		_, err := io.Copy(partB, partA)
		if err != nil {
			log.Fatal(err)
		}
		partB.Close()
	}()
	wg.Wait()
}

func open(a string) (io.ReadWriteCloser, error) {
	switch {
	case strings.HasPrefix(a, "TCP-LISTEN:"):
		ls, err := net.Listen("tcp", ":"+a[len("TCP-LISTEN:"):])
		if err != nil {
			return nil, err
		}
		conn, err := ls.Accept()
		if err != nil {
			return nil, err
		}
		return conn, nil
	case strings.HasPrefix(a, "UNIX-CONNECT:"):
		file := a[len("UNIX-CONNECT:"):]
		return net.Dial("unix", file)
	case a == "STDIO":
		return &stdIO{}, nil
	}
	return nil, fmt.Errorf("unsupported protocol: %s", a)
}

type stdIO struct{}

func (s *stdIO) Read(p []byte) (n int, err error) {
	return os.Stdin.Read(p)
}

func (s *stdIO) Write(p []byte) (n int, err error) {
	return os.Stdin.Write(p)
}

func (s *stdIO) Close() error {
	return nil
}
