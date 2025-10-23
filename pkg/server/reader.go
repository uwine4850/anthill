package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"

	"github.com/uwine4850/anthill/internal/pathutils"
)

const MAX_HISTORY_LEN = 300

type AntWorkerStreamer struct {
	Name    string
	logs    chan string
	history []string
	socket  string
	mu      sync.Mutex
}

func NewAntWorkerReader(antWorkerName string) *AntWorkerStreamer {
	return &AntWorkerStreamer{
		Name:    antWorkerName,
		history: make([]string, 0, MAX_HISTORY_LEN),
		logs:    make(chan string, 1),
		socket:  makeStreamSocket(antWorkerName),
	}
}

func (s *AntWorkerStreamer) ReadText(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		text := scanner.Text()
		s.mu.Lock()
		if len(s.history) > MAX_HISTORY_LEN {
			copy(s.history, s.history[1:])
			s.history[len(s.history)-1] = text
		} else {
			s.history = append(s.history, text)
		}
		s.mu.Unlock()
		s.logs <- text
	}
	close(s.logs)
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

			// Clearing recorded channels before connecting to avoid duplication with history
			s.drain()

			go func(c net.Conn) {
				defer c.Close()

				s.mu.Lock()
				for i := 0; i < len(s.history); i++ {
					if _, err := fmt.Fprintln(c, s.history[i]); err != nil {
						return
					}
				}
				s.mu.Unlock()

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

func (s *AntWorkerStreamer) drain() {
	for {
		select {
		case <-s.logs:
		default:
			return
		}
	}
}

func ReadStream(antWorkerName string) {
	conn, err := net.Dial("unix", makeStreamSocket(antWorkerName))
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

func makeStreamSocket(antWorkerName string) string {
	return fmt.Sprintf("/tmp/anthill-%s.sock", antWorkerName)
}
