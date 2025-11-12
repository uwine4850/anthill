package worker

import (
	"fmt"

	"github.com/uwine4850/anthill/pkg/domain"
	"github.com/uwine4850/anthill/pkg/infra/parsecnf"
)

func CurrentAnts1(workersConfig *parsecnf.WorkersConfig, allPluginAnts map[string]domain.PluginAnt) (map[string]domain.PluginAnt, error) {
	currentAnts := map[string]domain.PluginAnt{}
	for i := 0; i < len(workersConfig.Workers); i++ {
		workerConfig := workersConfig.Workers[i]
		if pluginAnt, ok := allPluginAnts[workerConfig.Type]; ok {
			pluginAnt.Args = workerConfig.Args
			pluginAnt.Reload = workerConfig.Reload
			pluginAnt.After = workerConfig.After
			currentAnts[workerConfig.Name] = pluginAnt
		} else {
			return nil, fmt.Errorf("WorkerAnt for type %s not found", workerConfig.Type)
		}
	}
	return currentAnts, nil
}
