package worker

import (
	"fmt"

	dmnworker "github.com/uwine4850/anthill/pkg/domain/dmn_worker"
	"github.com/uwine4850/anthill/pkg/infra/parsecnf"
)

func CurrentAnts(workersConfig *parsecnf.WorkersConfig, allPluginAnts map[string]dmnworker.PluginAnt) (map[string]dmnworker.PluginAnt, error) {
	currentAnts := map[string]dmnworker.PluginAnt{}
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
