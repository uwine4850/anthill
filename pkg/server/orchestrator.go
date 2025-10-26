package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/uwine4850/anthill/internal/pathutils"
	"github.com/uwine4850/anthill/pkg/config"
	"github.com/uwine4850/anthill/pkg/worker"
)

type AfterWorker struct {
	Conn      net.Conn
	PluginAnt worker.PluginAnt
	Request   worker.Request
}

type Orchestrator struct {
	currentAnts          map[string]worker.PluginAnt
	workersConfig        *config.WorkersConfig
	status               *worker.Status
	antWorkerProcess     AWorkerProcess
	startAfterWorkerAnts sync.Map
}

func NewOrchestartor() Orchestrator {
	return Orchestrator{
		currentAnts:          make(map[string]worker.PluginAnt, 0),
		status:               worker.NewStatus(),
		antWorkerProcess:     &AntWorkerProcess{},
		startAfterWorkerAnts: sync.Map{},
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

	go o.runDependentWorkers()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go func() {
			var req worker.Request
			decoder := json.NewDecoder(conn)
			if err := decoder.Decode(&req); err != nil {
				log.Printf("decode error: %s\n", err)
			}
			if len(o.currentAnts[req.Name].After) != 0 {
				o.startAfterWorkerAnts.Store(conn, AfterWorker{
					Conn:      conn,
					PluginAnt: o.currentAnts[req.Name],
					Request:   req,
				})
				return
			}
			if err := o.handleConnection(conn, req); err != nil {
				log.Printf("handle connection error: %s\n", err)
			}
		}()
	}
}

func (o *Orchestrator) handleConnection(conn net.Conn, req worker.Request) error {
	defer conn.Close()
	p := o.antWorkerProcess.New(&o.currentAnts, req.Name)
	p.OnDone(func() {
		if err := o.status.SetDone(req.Name); err != nil {
			log.Println(err)
		}
	})

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

func (o *Orchestrator) runDependentWorkers() {
	for {
		mustRunWorkers := []AfterWorker{}
		o.startAfterWorkerAnts.Range(func(key, value any) bool {
			afterWorker := value.(AfterWorker)
			isAllDone := false
			for i := 0; i < len(afterWorker.PluginAnt.After); i++ {
				s, err := o.status.Get(afterWorker.PluginAnt.After[i])
				if err != nil {
					log.Println(err)
				}
				if s.Done {
					isAllDone = true
				} else {
					isAllDone = false
				}
			}
			if isAllDone {
				mustRunWorkers = append(mustRunWorkers, afterWorker)
				return false
			}
			return true
		})
		for i := 0; i < len(mustRunWorkers); i++ {
			go func() {
				if err := o.handleConnection(mustRunWorkers[i].Conn, mustRunWorkers[i].Request); err != nil {
					log.Printf("handle connection error: %s", err)
				}
			}()
			o.startAfterWorkerAnts.Delete(mustRunWorkers[i].Conn)
		}
		time.Sleep(500 * time.Millisecond)
	}
}
