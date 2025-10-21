package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/uwine4850/anthill/internal/pathutils"
	"github.com/uwine4850/anthill/pkg/config"
	"github.com/uwine4850/anthill/pkg/worker"
)

type Orchestrator struct {
	currentAnts   map[string]worker.PluginAnt
	workersConfig *config.WorkersConfig
	status        *worker.Status
}

func NewOrchestartor() Orchestrator {
	return Orchestrator{
		currentAnts: make(map[string]worker.PluginAnt, 0),
		status:      worker.NewStatus(),
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
		if err := runWorker(o.currentAnts, runningWorkers, req.Name); err != nil {
			return err
		}
		if err := o.status.SetRunning(req.Name); err != nil {
			return err
		}
	case "stop":
		if err := cancelWorker(runningWorkers, req.Name); err != nil {
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

func runWorker(ants map[string]worker.PluginAnt, runningWorkers *sync.Map, name string) error {
	pluginAnt, ok := ants[name]
	if !ok {
		return fmt.Errorf("cannot run worker <%s>; it does not exists", name)
	}
	go func(ant worker.PluginAnt) {
		cmd := exec.Command("./launcher", append([]string{ant.Path}, ant.Args...)...)
		cmdStdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Println("Stdout pipe error:", err)
			return
		}
		cmdStderr, err := cmd.StderrPipe()
		if err != nil {
			log.Println("Stderr pipe error:", err)
			return
		}
		if err := cmd.Start(); err != nil {
			log.Println("Start error: ", err)
			return
		}
		runningWorkers.Store(name, cmd)

		antWorkerReader := NewAntWorkerReader(name)
		go antWorkerReader.ReadText(cmdStdout)
		go antWorkerReader.ReadText(cmdStderr)
		if err := antWorkerReader.Stream(); err != nil {
			log.Println("stream error: ", err)
			return
		}

		if err := cmd.Wait(); err != nil {
			log.Println(name, "wait error:", err)
			return
		}
	}(pluginAnt)
	return nil
}

func cancelWorker(runningWorkers *sync.Map, name string) error {
	_cmd, ok := runningWorkers.Load(name)
	if !ok {
		return fmt.Errorf("running worker <%s> not exists", name)
	}
	cmd := _cmd.(*exec.Cmd)
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return err
	}
	runningWorkers.Delete(name)
	return nil
}
