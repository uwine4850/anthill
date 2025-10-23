package worker

import (
	"fmt"

	"github.com/uwine4850/anthill/pkg/config"
	"github.com/uwine4850/anthill/pkg/plug"
)

type WorkerAnt interface {
	Run() error
	Stop() error
	Type() string
	Info() string
	Args(args ...string) error
}

type PluginAnt struct {
	Path      string
	Reload    bool
	Args      []string
	WorkerAnt WorkerAnt
}

func ExtractPluginAntsFromPlugins(pluginsConfig config.PluginsConfig) (map[string]PluginAnt, error) {
	builtinList, err := plug.BuiltinList()
	if err != nil {
		return nil, err
	}
	pluginsList := append(pluginsConfig.Plugins, builtinList...)

	pluginAnts := make(map[string]PluginAnt, len(pluginsList))
	for i := 0; i < len(pluginsList); i++ {
		pluginPath := pluginsList[i]
		workerAnt, err := WorkerAntFromPlugin(pluginPath)
		if err != nil {
			return nil, err
		}
		if _, ok := pluginAnts[workerAnt.Type()]; !ok {
			pluginAnts[workerAnt.Type()] = PluginAnt{
				Path:      pluginPath,
				WorkerAnt: workerAnt,
			}
		} else {
			return nil, fmt.Errorf("WorkerAnt type %s already exists", workerAnt.Type())
		}
	}
	return pluginAnts, nil
}

func CurrentAnts(workersConfig *config.WorkersConfig, allPluginAnts map[string]PluginAnt) (map[string]PluginAnt, error) {
	currentAnts := map[string]PluginAnt{}
	for i := 0; i < len(workersConfig.Workers); i++ {
		workerConfig := workersConfig.Workers[i]
		if pluginAnt, ok := allPluginAnts[workerConfig.Type]; ok {
			pluginAnt.Args = workerConfig.Args
			pluginAnt.Reload = workerConfig.Reload
			currentAnts[workerConfig.Name] = pluginAnt
		} else {
			return nil, fmt.Errorf("WorkerAnt for type %s not found", workerConfig.Type)
		}
	}
	return currentAnts, nil
}

func WorkerAntListFromPlugins(path string) (map[string]WorkerAnt, error) {
	pluginList, err := config.ParsePlugins(path)
	if err != nil {
		return nil, err
	}
	ants := make(map[string]WorkerAnt, len(pluginList.Plugins))
	for i := 0; i < len(pluginList.Plugins); i++ {
		workerAnt, err := WorkerAntFromPlugin(pluginList.Plugins[i])
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

func WorkerAntFromPlugin(path string) (WorkerAnt, error) {
	plugin, err := plug.OpenPlugin(path)
	if err != nil {
		return nil, err
	}
	workerAnt := (*plugin).(WorkerAnt)
	return workerAnt, nil
}
