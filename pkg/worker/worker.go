package worker

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"

	"github.com/uwine4850/anthill/pkg/config"
	"github.com/uwine4850/anthill/pkg/plug"
)

type WorkerAnt interface {
	Run() error
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

type Runner struct {
	stdout map[string]string
}

func (r *Runner) RunWorkers(ants []Ant) {
	var wg sync.WaitGroup
	for i := 0; i < len(ants); i++ {
		wg.Add(1)
		go func(ant Ant) {
			defer wg.Done()
			err := runPluginStreaming(ant.Name)
			if err != nil {
				panic(err)
			}
		}(ants[i])
	}
	wg.Wait()
}

func runPluginStreaming(antName string) error {
	cmd := exec.Command(os.Args[0], "run", antName)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	go streamOutput(stdoutPipe, antName)
	go streamOutput(stderrPipe, antName+"[ERR]")

	return cmd.Wait()
}

func streamOutput(r io.ReadCloser, prefix string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Printf("[%s] %s\n", prefix, scanner.Text())
	}
}

type Request struct {
	Action string `json:"action"`
	Name   string `json:"name"`
}

func RunAllWorkers() error {
	w, err := config.ParseWorkers("workers.yaml")
	if err != nil {
		return nil
	}

	var wg sync.WaitGroup
	for i := 0; i < len(w.Workers); i++ {
		wg.Add(1)
		go func(name string) {
			conn, err := net.Dial("unix", "/tmp/anthill.sock")
			if err != nil {
				log.Fatal("failed to connect to orchestrator:", err)
			}
			defer conn.Close()

			defer wg.Done()
			req := Request{Action: "run", Name: w.Workers[i].Name}
			enc := json.NewEncoder(conn)
			err = enc.Encode(req)
			if err != nil {
				log.Fatal("failed to send request:", err)
			}
		}(w.Workers[i].Name)
	}
	wg.Wait()
	return nil
}
