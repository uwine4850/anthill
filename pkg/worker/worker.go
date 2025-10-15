package worker

import (
	"fmt"

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
