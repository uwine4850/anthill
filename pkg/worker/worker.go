package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/uwine4850/anthill/pkg/config"
	"github.com/uwine4850/anthill/pkg/plug"
)

type WorkerAnt interface {
	Run(ctx context.Context) error
	Type() string
	Info() string
	Args(args map[string]string) error
}

type Ant struct {
	Name   string
	Reload bool
	Worker WorkerAnt
}

func WorkerAntListFromPlugins(path string) (map[string]WorkerAnt, error) {
	pluginList, err := config.ParsePlugins(path)
	if err != nil {
		return nil, err
	}
	ants := make(map[string]WorkerAnt, len(pluginList.Plugins))
	for i := 0; i < len(pluginList.Plugins); i++ {
		workerAnt, err := workerAntFromPlugin(pluginList.Plugins[i])
		if err != nil {
			return nil, err
		}
		if _, ok := ants[workerAnt.Type()]; ok {
			return nil, fmt.Errorf("WorkerAnt type <%s> already exists", workerAnt.Type())
		}
		ants[workerAnt.Type()] = workerAnt
	}
	return ants, nil
}

func workerAntFromPlugin(path string) (WorkerAnt, error) {
	plugin, err := plug.OpenPlugin(path)
	if err != nil {
		return nil, err
	}
	workerAnt := (*plugin).(WorkerAnt)
	return workerAnt, nil
}

func TotalWorkerAntList(pluginWorkerAnts map[string]WorkerAnt) (map[string]WorkerAnt, error) {
	for i := 0; i < len(builtinWorkerAnts); i++ {
		workerAnt := builtinWorkerAnts[i]
		if _, ok := pluginWorkerAnts[workerAnt.Type()]; ok {
			return nil, fmt.Errorf("WorkerAnt type <%s> already exists", workerAnt.Type())
		}
		pluginWorkerAnts[workerAnt.Type()] = workerAnt
	}
	return pluginWorkerAnts, nil
}

func CurrentAnts(workersConfig *config.WorkersConfig, allWorkers map[string]WorkerAnt) ([]Ant, error) {
	currentAnts := []Ant{}
	for i := 0; i < len(workersConfig.Workers); i++ {
		workerConfig := workersConfig.Workers[i]
		if workerAnt, ok := allWorkers[workerConfig.Type]; ok {
			if err := initWorkerAnt(workerAnt, &workerConfig); err != nil {
				return nil, err
			}
			currentAnts = append(currentAnts, Ant{
				Name:   workerConfig.Name,
				Reload: workerConfig.Reload,
				Worker: workerAnt,
			})
		} else {
			return nil, fmt.Errorf("WorkerAnt for type %s not found", workerConfig.Type)
		}
	}
	return currentAnts, nil
}

func initWorkerAnt(ant WorkerAnt, workerConfig *config.WorkerConfig) error {
	if err := ant.Args(workerConfig.Args); err != nil {
		return err
	}
	return nil
}

type Request struct {
	Action string `json:"action"`
	Name   string `json:"name"`
}

type Runner struct {
	workersPath   string
	workersConfig *config.WorkersConfig
	wg            sync.WaitGroup
}

func NewRunner(workersPath string) Runner {
	return Runner{
		workersPath: workersPath,
	}
}

func (r *Runner) Init() error {
	w, err := config.ParseWorkers("workers.yaml")
	if err != nil {
		return err
	}
	r.workersConfig = w
	return nil
}

func (r *Runner) Wait() {
	r.wg.Wait()
}

func (r *Runner) RunAllWorkers() error {
	workersConfig := r.workersConfig.Workers
	for i := 0; i < len(workersConfig); i++ {
		r.wg.Add(1)
		go func(name string) {
			defer r.wg.Done()
			conn, err := connectToOrchestrator()
			if err != nil {
				log.Fatal(err)
			}
			defer conn.Close()

			req := Request{Action: "run", Name: workersConfig[i].Name}
			enc := json.NewEncoder(conn)
			err = enc.Encode(req)
			if err != nil {
				log.Fatal("failed to send request:", err)
			}
		}(workersConfig[i].Name)
	}
	return nil
}

func (r *Runner) RunWorker(name string) {
	for i := 0; i < len(r.workersConfig.Workers); i++ {
		workerConfig := r.workersConfig.Workers[i]
		if workerConfig.Name == name {
			r.wg.Add(1)
			go func() {
				defer r.wg.Done()
				conn, err := connectToOrchestrator()
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				req := Request{Action: "run", Name: name}
				enc := json.NewEncoder(conn)
				err = enc.Encode(req)
				if err != nil {
					log.Fatal("failed to send request:", err)
				}
			}()
		}
	}
}

func StopWorker(name string) error {
	conn, err := connectToOrchestrator()
	if err != nil {
		return err
	}
	defer conn.Close()

	req := Request{Action: "stop", Name: name}
	enc := json.NewEncoder(conn)
	err = enc.Encode(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %s", err)
	}
	return nil
}

func connectToOrchestrator() (net.Conn, error) {
	conn, err := net.Dial("unix", "/tmp/anthill.sock")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to orchestrator: %s", err)
	}
	return conn, nil
}
