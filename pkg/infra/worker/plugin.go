package worker

import (
	"fmt"

	dmnworker "github.com/uwine4850/anthill/pkg/domain/dmn_worker"
	"github.com/uwine4850/anthill/pkg/infra/plug"
)

func ExtractPluginAntsFromPlugins(pluginsConfig dmnworker.PluginsConfig) (map[string]dmnworker.PluginAnt, error) {
	builtinList, err := plug.BuiltinList()
	if err != nil {
		return nil, err
	}
	pluginsList := append(pluginsConfig.Plugins, builtinList...)

	pluginAnts := make(map[string]dmnworker.PluginAnt, len(pluginsList))
	for i := 0; i < len(pluginsList); i++ {
		pluginPath := pluginsList[i]
		workerAnt, err := WorkerAntFromPlugin(pluginPath)
		if err != nil {
			return nil, err
		}
		if _, ok := pluginAnts[workerAnt.Type()]; !ok {
			pluginAnts[workerAnt.Type()] = dmnworker.PluginAnt{
				Path:      pluginPath,
				WorkerAnt: workerAnt,
			}
		} else {
			return nil, fmt.Errorf("WorkerAnt type %s already exists", workerAnt.Type())
		}
	}
	return pluginAnts, nil
}

func WorkerAntFromPlugin(path string) (dmnworker.WorkerAnt, error) {
	plugin, err := plug.OpenPlugin(path)
	if err != nil {
		return nil, err
	}
	workerAnt := (*plugin).(dmnworker.WorkerAnt)
	return workerAnt, nil
}
