package worker

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
