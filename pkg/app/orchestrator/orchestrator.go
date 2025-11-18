package orchestrator

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
	dmnsocket "github.com/uwine4850/anthill/pkg/domain/dmn_socket"
	dmnworker "github.com/uwine4850/anthill/pkg/domain/dmn_worker"
	"github.com/uwine4850/anthill/pkg/infra/parsecnf"
	"github.com/uwine4850/anthill/pkg/infra/process"
	"github.com/uwine4850/anthill/pkg/infra/status"
	"github.com/uwine4850/anthill/pkg/infra/worker"
)

type afterWorker struct {
	Conn      net.Conn
	PluginAnt dmnworker.PluginAnt
	Request   dmnsocket.Request
}

type Orchestrator struct {
	currentAnts          map[string]dmnworker.PluginAnt
	workersConfig        *parsecnf.WorkersConfig
	status               status.Status
	antWorkerProcess     dmnworker.AWorkerProcess
	startAfterWorkerAnts sync.Map
}

func NewOrchestartor() Orchestrator {
	return Orchestrator{
		currentAnts:          make(map[string]dmnworker.PluginAnt, 0),
		status:               status.NewStatus(),
		antWorkerProcess:     &process.AntWorkerProcess{},
		startAfterWorkerAnts: sync.Map{},
	}
}

func (o *Orchestrator) initStatus() {
	o.status.Init(o.workersConfig)
}

func (o *Orchestrator) CollectAnts() error {
	plugs, err := parsecnf.ParsePlugins("plugins.yaml")
	if err != nil {
		return err
	}
	pluginAnts, err := worker.ExtractPluginAntsFromPlugins(*plugs)
	if err != nil {
		return err
	}
	workersc, err := parsecnf.ParseWorkers("workers.yaml")
	if err != nil {
		return err
	}
	o.workersConfig = workersc

	currentAnts, err := worker.CurrentAnts(workersc, pluginAnts)
	if err != nil {
		return err
	}
	o.currentAnts = currentAnts
	return nil
}

func (o *Orchestrator) validateAnthillSocketPath() error {
	if err := pathutils.Exists(config.ANTHILL_SOCKET_PATH); err == nil {
		if err := os.Remove(config.ANTHILL_SOCKET_PATH); err != nil {
			return err
		}
	}
	return nil
}

func (o *Orchestrator) Listen() error {
	if err := o.validateAnthillSocketPath(); err != nil {
		return err
	}
	o.initStatus()

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
			var req dmnsocket.Request
			decoder := json.NewDecoder(conn)
			if err := decoder.Decode(&req); err != nil {
				log.Printf("decode error: %s\n", err)
			}
			if len(o.currentAnts[req.Name].After) != 0 {
				o.startAfterWorkerAnts.Store(conn, afterWorker{
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

func (o *Orchestrator) handleConnection(conn net.Conn, req dmnsocket.Request) error {
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
		if req.Name == "" {
			if err := status.SendResponse(conn, o.status); err != nil {
				return err
			}
		} else {
			if err := status.SendWorkerResponse(conn, req.Name, o.status); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("undefined action <%s>", req.Action)
	}
	return nil
}

func (o *Orchestrator) runDependentWorkers() {
	for {
		mustRunWorkers := []afterWorker{}
		o.startAfterWorkerAnts.Range(func(key, value any) bool {
			afterWorker := value.(afterWorker)
			isAllDone := false
			for i := 0; i < len(afterWorker.PluginAnt.After); i++ {
				s, ok := o.status.Get()[afterWorker.PluginAnt.After[i]]
				if !ok {
					log.Println(fmt.Errorf("running worker <%s> not exists", afterWorker.PluginAnt.After[i]))
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
