package main

import (
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"github.com/uwine4850/anthill/pkg/plug"
)

func main() {
	plugin, err := plug.OpenPlugin(os.Args[1])
	if err != nil {
		panic(err)
	}
	v := reflect.ValueOf(*plugin)
	runMethod := v.MethodByName("Run")
	stopMethod := v.MethodByName("Stop")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM)

	go func() {
		<-sigs
		stopMethod.Call(nil)
		os.Exit(0)
	}()
	runMethod.Call(nil)
}
