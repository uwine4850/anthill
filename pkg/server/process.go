package server

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"syscall"

	"github.com/uwine4850/anthill/pkg/worker"
)

type AWorkerProcess interface {
	Run() error
	Stop() error
	New(ants *map[string]worker.PluginAnt, name string) AWorkerProcess
}

type AntWorkerProcess struct {
	ants           *map[string]worker.PluginAnt
	runningWorkers *sync.Map
	name           string
	streamer       Streamer
}

func (p *AntWorkerProcess) New(ants *map[string]worker.PluginAnt, name string) AWorkerProcess {
	return &AntWorkerProcess{
		ants:           ants,
		runningWorkers: &sync.Map{},
		name:           name,
	}
}

func (p *AntWorkerProcess) Run() error {
	pluginAnt, ok := (*p.ants)[p.name]
	if !ok {
		return fmt.Errorf("cannot run worker <%s>; it does not exists", p.name)
	}
	go func(ant worker.PluginAnt) {
		cmd, stdout, stderr, err := p.initLauncher(&ant)
		if err != nil {
			log.Println(err)
		}
		if err := cmd.Start(); err != nil {
			log.Println("Start error: ", err)
			return
		}
		p.runningWorkers.Store(p.name, cmd)

		if err := p.initAndRunStreamer(stdout, stderr); err != nil {
			log.Println(err)
		}

		if err := cmd.Wait(); err != nil {
			log.Println(p.name, "wait error:", err)
			go func() {
				if err := p.killAndReloadOnError(pluginAnt); err != nil {
					log.Printf("Relaod %s error: %s\n", p.name, err)
				}
			}()
			return
		}
	}(pluginAnt)
	return nil
}

func (p *AntWorkerProcess) Stop() error {
	_cmd, ok := p.runningWorkers.Load(p.name)
	if !ok {
		return fmt.Errorf("running worker <%s> not exists", p.name)
	}
	cmd := _cmd.(*exec.Cmd)
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return err
	}
	p.runningWorkers.Delete(p.name)
	return nil
}

func (p *AntWorkerProcess) initLauncher(pluginAnt *worker.PluginAnt) (cmd *exec.Cmd, stdout io.Reader, stderr io.Reader, err error) {
	cmd = exec.Command("./launcher", append([]string{pluginAnt.Path}, pluginAnt.Args...)...)
	cmdStdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("stdout pipe error: %s", err)
	}
	cmdStderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("stderr pipe error: %s", err)
	}
	return cmd, cmdStdout, cmdStderr, nil
}

func (p *AntWorkerProcess) initAndRunStreamer(stdout io.Reader, stderr io.Reader) error {
	streamer := p.streamer.New(p.name)
	go streamer.ReadText(stdout)
	go streamer.ReadText(stderr)

	if err := streamer.Stream(); err != nil {
		return fmt.Errorf("stream error: %s", err)
	}
	return nil
}

func (p *AntWorkerProcess) killAndReloadOnError(pluginAnt worker.PluginAnt) error {
	if pluginAnt.Reload {
		if err := p.kill(); err != nil {
			return err
		}
		if err := p.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (p *AntWorkerProcess) kill() error {
	_cmd, ok := p.runningWorkers.Load(p.name)
	if !ok {
		return fmt.Errorf("running worker <%s> not exists", p.name)
	}
	cmd := _cmd.(*exec.Cmd)
	if cmd.ProcessState == nil || !cmd.ProcessState.Exited() {
		if err := cmd.Process.Kill(); err != nil {
			return err
		}
	}
	p.runningWorkers.Delete(p.name)
	return nil
}
