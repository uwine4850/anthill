package worker

import (
	"fmt"
	"time"
)

var builtinWorkerAnts = []WorkerAnt{
	&Command{},
}

type Command struct{}

func (c *Command) Run() error {
	time.Sleep(2 * time.Second)
	fmt.Println("Run... builtin")
	time.Sleep(2 * time.Second)
	fmt.Println("Run...1 builtin")
	time.Sleep(2 * time.Second)
	fmt.Println("Run...2 builtin")
	time.Sleep(2 * time.Second)
	fmt.Println("Run...3 builtin")
	return nil
}

func (c *Command) Type() string {
	return "tee"
}

func (c *Command) Info() string {
	return "Command worker"
}

func (c *Command) Args(values map[string]string) error {
	return nil
}
