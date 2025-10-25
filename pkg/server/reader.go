package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"sync/atomic"

	"github.com/uwine4850/anthill/internal/pathutils"
)

type Streamer interface {
	io.Closer
	ReadText(reader io.Reader)
	Stream() error
}

const MAX_HISTORY_LEN = 300

type AntWorkerStreamer struct {
	Name     string
	logs     chan string
	history  []string
	socket   string
	mu       sync.Mutex
	listener net.Listener
	isClose  atomic.Bool
	wg       sync.WaitGroup
}

func NewAntWorkerStreamer(antWorkerName string) Streamer {
	return &AntWorkerStreamer{
		Name:    antWorkerName,
		history: make([]string, 0, MAX_HISTORY_LEN),
		logs:    make(chan string, 1),
		socket:  makeStreamSocket(antWorkerName),
	}
}

func (s *AntWorkerStreamer) Close() error {
	s.isClose.Store(true)
	s.wg.Wait()
	close(s.logs)
	return s.listener.Close()
}

func (s *AntWorkerStreamer) ReadText(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	s.wg.Add(1)
	for scanner.Scan() && !s.isClose.Load() {
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
		s.wg.Done()
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
	s.listener = listener

	go func() {
		defer listener.Close()
		for !s.isClose.Load() {
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
