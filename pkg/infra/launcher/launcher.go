package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/uwine4850/anthill/pkg/infra/worker"
)

func main() {
	workerAnt, err := worker.WorkerAntFromPlugin(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM)

	go func() {
		<-sigs
		if err := workerAnt.Stop(); err != nil {
			log.Fatalln(err)
		}
		os.Exit(0)
	}()
	if len(os.Args) > 2 {
		if err := workerAnt.Args(os.Args[2:]...); err != nil {
			log.Fatalln(err)
		}
	}
	if err := workerAnt.Run(); err != nil {
		log.Fatalln(err)
	}
}
