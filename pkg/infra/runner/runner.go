package runner

import (
	"fmt"
	"log"
	"sync"

	"github.com/uwine4850/anthill/pkg/infra/parsecnf"
	"github.com/uwine4850/anthill/pkg/infra/socket"
	"github.com/uwine4850/anthill/pkg/infra/status"
)

type Runner struct {
	workersPath   string
	workersConfig *parsecnf.WorkersConfig
	wg            sync.WaitGroup
}

func NewRunner(workersPath string) Runner {
	return Runner{
		workersPath: workersPath,
	}
}

func (r *Runner) Init() error {
	w, err := parsecnf.ParseWorkers("workers.yaml")
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
		workerStatus, err := status.CheckStatus(workersConfig[i].Name)
		if err != nil {
			return err
		}
		if workerStatus.Active {
			fmt.Printf("worker %s already active\n", workersConfig[i].Name)
		}
		r.wg.Add(1)
		go func(name string) {
			defer r.wg.Done()
			conn, err := socket.ConnectToOrchestrator()
			if err != nil {
				log.Fatal(err)
			}
			defer conn.Close()

			req := socket.Request{Action: "run", Name: workersConfig[i].Name}
			if err := socket.SendRequest(conn, req); err != nil {
				log.Fatal("failed to send request:", err)
			}
		}(workersConfig[i].Name)
	}
	return nil
}

func (r *Runner) RunWorker(name string) error {
	workerStatus, err := status.CheckStatus(name)
	if err != nil {
		return err
	}
	if workerStatus.Active {
		fmt.Printf("worker %s already active\n", name)
	}
	for i := 0; i < len(r.workersConfig.Workers); i++ {
		workerConfig := r.workersConfig.Workers[i]
		if workerConfig.Name == name {
			r.wg.Add(1)
			go func() {
				defer r.wg.Done()
				conn, err := socket.ConnectToOrchestrator()
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()

				req := socket.Request{Action: "run", Name: name}
				if err := socket.SendRequest(conn, req); err != nil {
					log.Fatal("failed to send request:", err)
				}
			}()
		} else {
			log.Fatalf("worker <%s> not exists\n", name)
		}
	}
	return err
}

func (r *Runner) StopWorker(name string) error {
	conn, err := socket.ConnectToOrchestrator()
	if err != nil {
		return err
	}
	defer conn.Close()

	for i := 0; i < len(r.workersConfig.Workers); i++ {
		workerConfig := r.workersConfig.Workers[i]
		if workerConfig.Name == name {
			req := socket.Request{Action: "stop", Name: name}
			if err := socket.SendRequest(conn, req); err != nil {
				log.Fatal("failed to send request:", err)
			}
		} else {
			return fmt.Errorf("worker <%s> not exists", name)
		}
	}
	return nil
}
