package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/uwine4850/anthill/internal/pathutils"
	"github.com/uwine4850/anthill/pkg/config"
	"github.com/uwine4850/anthill/pkg/worker"
)

type Orchestrator struct {
	currentAnts []worker.Ant
}

func NewOrchestartor() Orchestrator {
	return Orchestrator{
		currentAnts: make([]worker.Ant, 0),
	}
}

func (o *Orchestrator) CollectAnts() error {
	ants, err := worker.WorkerAntListFromPlugins("plugins.yaml")
	if err != nil {
		return err
	}
	a, err := worker.TotalWorkerAntList(ants)
	if err != nil {
		return err
	}

	w, err := config.ParseWorkers("workers.yaml")
	if err != nil {
		return err
	}

	wcurrent, err := worker.CurrentAnts(w, a)
	if err != nil {
		return err
	}

	o.currentAnts = wcurrent
	return nil
}

func (o *Orchestrator) Listen() error {
	socketPath := "/tmp/anthill.sock"
	if err := pathutils.Exists(socketPath); err == nil {
		if err := os.Remove(socketPath); err != nil {
			return err
		}
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Println("Online.")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go func() {
			if err := o.handleConnection(conn); err != nil {
				log.Fatalf("fatal error: %s", err)
			}
		}()
	}
}

func (o *Orchestrator) handleConnection(conn net.Conn) error {
	defer conn.Close()

	var req worker.Request
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&req); err != nil {
		return err
	}

	if req.Action == "run" {
		for i := 0; i < len(o.currentAnts); i++ {
			if o.currentAnts[i].Name == req.Name {
				if err := o.currentAnts[i].Worker.Run(); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
