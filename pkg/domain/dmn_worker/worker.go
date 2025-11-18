package dmnworker

type WorkerAnt interface {
	Run() error
	Stop() error
	Type() string
	Info() string
	Args(args ...string) error
}

type WorkerConfig struct {
	Name   string
	Reload bool
	Type   string
	After  []string
	Args   []string
}

type AWorkerProcess interface {
	Run() error
	Stop() error
	OnDone(fn func())
	New(ants *map[string]PluginAnt, name string) AWorkerProcess
}
