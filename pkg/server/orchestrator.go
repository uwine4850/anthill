package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

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

	runningWorkers := sync.Map{}

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go func() {
			if err := o.handleConnection(conn, &runningWorkers); err != nil {
				log.Fatalf("fatal error: %s", err)
			}
		}()
	}
}

func (o *Orchestrator) handleConnection(conn net.Conn, runningWorkers *sync.Map) error {
	defer conn.Close()

	var req worker.Request
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&req); err != nil {
		return err
	}
	switch req.Action {
	case "run":
		runWorker(o.currentAnts, runningWorkers, req.Name)
	case "stop":
		if err := cancelWorker(runningWorkers, req.Name); err != nil {
			return nil
		}
	default:
		return fmt.Errorf("undefined action <%s>", req.Action)
	}
	return nil
}

func runWorker(ants []worker.Ant, runningWorkers *sync.Map, name string) {
	for i := 0; i < len(ants); i++ {
		if ants[i].Name == name {
			ctx, cancel := context.WithCancel(context.Background())
			runningWorkers.Store(name, cancel)
			go func(ant worker.Ant, ctx context.Context) {
				if err := ant.Worker.Run(ctx); err != nil {
					log.Printf("worker error: %v", err)
				}
			}(ants[i], ctx)
		}
	}
}

func cancelWorker(runningWorkers *sync.Map, name string) error {
	cancelFunc, ok := runningWorkers.Load(name)
	if !ok {
		return fmt.Errorf("running worker <%s> not exists", name)
	}
	cancel := cancelFunc.(context.CancelFunc)
	cancel()
	runningWorkers.Delete(name)
	return nil
}
