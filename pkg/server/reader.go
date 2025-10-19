package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/uwine4850/anthill/internal/pathutils"
)

type AntWorkerStreamer struct {
	Name   string
	logs   chan string
	socket string
}

func NewAntWorkerReader(antWorkerName string) *AntWorkerStreamer {
	return &AntWorkerStreamer{
		Name:   antWorkerName,
		logs:   make(chan string, 1),
		socket: fmt.Sprintf("/tmp/anthill-%s.sock", antWorkerName),
	}
}

func (s *AntWorkerStreamer) ReadText(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		s.logs <- scanner.Text()
	}
}

func (s *AntWorkerStreamer) Stream() error {
	if err := pathutils.Exists(s.socket); err == nil {
		if err := os.Remove(s.socket); err != nil {
			return err
		}
	}

	listener, err := net.Listen("unix", s.socket)
	if err != nil {
		return err
	}

	go func() {
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("socket accept error:", err)
				continue
			}
			go func(c net.Conn) {
				defer c.Close()
				for lineCh := range s.logs {
					if _, err := fmt.Fprintln(c, lineCh); err != nil {
						return
					}
				}
			}(conn)
		}
	}()
	return nil
}

func ReadStream(socketPath string) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		log.Fatalf("failed connect to socket: %v", err)
	}
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("read error: %v", err)
	}
}
