package main

import (
	"fmt"
	"time"

	_ "github.com/uwine4850/anthill/pkg/worker"
)

type Command struct{}

func (c *Command) Run() error {
	var i int
	for {
		time.Sleep(4 * time.Second)
		fmt.Println("Run... cmd", i)
		i++
	}
}

func (c *Command) Stop() error {
	fmt.Println("STOP CMD")
	return nil
}

func (c *Command) Type() string {
	return "cmd"
}

func (c *Command) Info() string {
	fmt.Println("INFO")
	return "Command worker"
}

func (c *Command) Args(args ...any) error {
	fmt.Println("Cmd args:", args)
	return nil
}

var Plugin Command
