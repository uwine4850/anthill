package main

import "github.com/uwine4850/anthill/pkg/server"

func main() {
	o := server.NewOrchestartor()
	if err := o.CollectAnts(); err != nil {
		panic(err)
	}
	if err := o.Listen(); err != nil {
		panic(err)
	}
}
