package dmnworker

type PluginsConfig struct {
	Plugins []string
}

type PluginAnt struct {
	Path      string
	Reload    bool
	After     []string
	Args      []string
	WorkerAnt WorkerAnt
}
