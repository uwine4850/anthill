package worker

import (
	"fmt"

	"github.com/uwine4850/anthill/pkg/domain"
	"github.com/uwine4850/anthill/pkg/infra/plug"
)

func ExtractPluginAntsFromPlugins(pluginsConfig domain.PluginsConfig) (map[string]domain.PluginAnt, error) {
	builtinList, err := plug.BuiltinList()
	if err != nil {
		return nil, err
	}
	pluginsList := append(pluginsConfig.Plugins, builtinList...)

	pluginAnts := make(map[string]domain.PluginAnt, len(pluginsList))
	for i := 0; i < len(pluginsList); i++ {
		pluginPath := pluginsList[i]
		workerAnt, err := WorkerAntFromPlugin(pluginPath)
		if err != nil {
			return nil, err
		}
		if _, ok := pluginAnts[workerAnt.Type()]; !ok {
			pluginAnts[workerAnt.Type()] = domain.PluginAnt{
				Path:      pluginPath,
				WorkerAnt: workerAnt,
			}
		} else {
			return nil, fmt.Errorf("WorkerAnt type %s already exists", workerAnt.Type())
		}
	}
	return pluginAnts, nil
}

func WorkerAntFromPlugin(path string) (domain.WorkerAnt, error) {
	plugin, err := plug.OpenPlugin(path)
	if err != nil {
		return nil, err
	}
	workerAnt := (*plugin).(domain.WorkerAnt)
	return workerAnt, nil
}
