package main

import (
	"fmt"

	"github.com/uwine4850/anthill/pkg/config"
	"github.com/uwine4850/anthill/pkg/worker"
)

func main() {
	ants, err := worker.WorkerAntListFromPlugins("plugins.yaml")
	if err != nil {
		panic(err)
	}
	a, err := worker.TotalWorkerAntList(ants)
	if err != nil {
		panic(err)
	}

	w, err := config.ParseWorkers("workers.yaml")
	if err != nil {
		panic(err)
	}

	wcurrent, err := worker.CurrentAnts(w, a)
	if err != nil {
		panic(err)
	}
	fmt.Println(wcurrent)
}
