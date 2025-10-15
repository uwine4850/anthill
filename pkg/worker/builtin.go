package worker

import "fmt"

var builtinWorkerAnts = []WorkerAnt{
	&Command{},
}

type Command struct{}

func (c *Command) Run() error {
	fmt.Println("Run...")
	return nil
}

func (c *Command) Type() string {
	return "tee"
}

func (c *Command) Info() string {
	return "Command worker"
}

func (c *Command) Args(values map[string]string) error {
	fmt.Println("Args:", values)
	return nil
}
