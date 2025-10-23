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
	currentAnts      map[string]worker.PluginAnt
	workersConfig    *config.WorkersConfig
	status           *worker.Status
	antWorkerProcess AWorkerProcess
}

func NewOrchestartor() Orchestrator {
	return Orchestrator{
		currentAnts:      make(map[string]worker.PluginAnt, 0),
		status:           worker.NewStatus(),
		antWorkerProcess: &AntWorkerProcess{},
	}
}

func (o *Orchestrator) CollectAnts() error {
	plugs, err := config.ParsePlugins("plugins.yaml")
	if err != nil {
		return err
	}
	pluginAnts, err := worker.ExtractPluginAntsFromPlugins(*plugs)
	if err != nil {
		return err
	}
	workersc, err := config.ParseWorkers("workers.yaml")
	if err != nil {
		return err
	}
	o.workersConfig = workersc
	o.status.Init(workersc)

	currentAnts, err := worker.CurrentAnts(workersc, pluginAnts)
	if err != nil {
		return err
	}
	o.currentAnts = currentAnts
	return nil
}

func (o *Orchestrator) Listen() error {
	if err := pathutils.Exists(config.ANTHILL_SOCKET_PATH); err == nil {
		if err := os.Remove(config.ANTHILL_SOCKET_PATH); err != nil {
			return err
		}
	}
	listener, err := net.Listen("unix", config.ANTHILL_SOCKET_PATH)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Println("Orchestrator online.")

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

	p := o.antWorkerProcess.New(&o.currentAnts, req.Name)

	switch req.Action {
	case "run":
		if err := p.Run(); err != nil {
			return err
		}
		if err := o.status.SetRunning(req.Name); err != nil {
			return err
		}
	case "stop":
		if err := p.Stop(); err != nil {
			return err
		}
		if err := o.status.SetStopped(req.Name); err != nil {
			return err
		}
	case "status":
		if err := o.status.SendResponse(conn); err != nil {
			return err
		}
	default:
		return fmt.Errorf("undefined action <%s>", req.Action)
	}
	return nil
}
