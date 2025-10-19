package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/uwine4850/anthill/pkg/worker"
)

func main() {
	workerAnt, err := worker.WorkerAntFromPlugin(os.Args[1])
	if err != nil {
		panic(err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM)

	go func() {
		<-sigs
		if err := workerAnt.Stop(); err != nil {
			panic(err)
		}
		os.Exit(0)
	}()
	if len(os.Args) > 2 {
		if err := workerAnt.Args(os.Args[2:]...); err != nil {
			panic(err)
		}
	}
	if err := workerAnt.Run(); err != nil {
		panic(err)
	}
}
