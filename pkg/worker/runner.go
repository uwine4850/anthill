package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/uwine4850/anthill/pkg/config"
)

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
		} else {
			log.Fatalf("worker <%s> not exists\n", name)
		}
	}
}

func (r *Runner) StopWorker(name string) error {
	conn, err := connectToOrchestrator()
	if err != nil {
		return err
	}
	defer conn.Close()

	for i := 0; i < len(r.workersConfig.Workers); i++ {
		workerConfig := r.workersConfig.Workers[i]
		if workerConfig.Name == name {
			req := Request{Action: "stop", Name: name}
			enc := json.NewEncoder(conn)
			err = enc.Encode(req)
			if err != nil {
				return fmt.Errorf("failed to send request: %s", err)
			}
		} else {
			return fmt.Errorf("worker <%s> not exists", name)
		}
	}
	return nil
}

func connectToOrchestrator() (net.Conn, error) {
	conn, err := net.Dial("unix", config.ANTHILL_SOCKET_PATH)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to orchestrator: %s", err)
	}
	return conn, nil
}
