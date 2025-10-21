package worker

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/uwine4850/anthill/pkg/config"
)

type StatusResponse struct {
	Name   string
	Active bool
	UpDate time.Time
}

type Status struct {
	workerAntsStatus map[string]StatusResponse
}

func NewStatus() *Status {
	return &Status{
		workerAntsStatus: make(map[string]StatusResponse),
	}
}

func (s *Status) Init(workersConfig *config.WorkersConfig) {
	for i := 0; i < len(workersConfig.Workers); i++ {
		w := workersConfig.Workers[i]
		s.workerAntsStatus[w.Name] = StatusResponse{
			Name: w.Name,
		}
	}
}

func (s *Status) SetRunning(name string) error {
	w, ok := s.workerAntsStatus[name]
	if ok {
		w.Active = true
		w.UpDate = time.Now()
		s.workerAntsStatus[name] = w
	} else {
		return fmt.Errorf("worker %s not exists", name)
	}
	return nil
}

func (s *Status) SetStopped(name string) error {
	w, ok := s.workerAntsStatus[name]
	if ok {
		w.Active = false
		w.UpDate = time.Time{}
		s.workerAntsStatus[name] = w
	} else {
		return fmt.Errorf("worker %s not exists", name)
	}
	return nil
}

func (s *Status) SendResponse(conn net.Conn) error {
	data, err := json.Marshal(s.workerAntsStatus)
	if err != nil {
		return err
	}
	if _, err := conn.Write(data); err != nil {
		return err
	}
	return nil
}

func CheckStatus() error {
	conn, err := connectToOrchestrator()
	if err != nil {
		return err
	}

	req := Request{Action: "status"}
	enc := json.NewEncoder(conn)
	err = enc.Encode(req)
	if err != nil {
		return err
	}

	var resp map[string]WorkerStatus
	dec := json.NewDecoder(conn)
	if err := dec.Decode(&resp); err != nil {
		return err
	}
	for _, status := range resp {
		f, err := fmt.Printf("Name: %s | Active: %v, | UpDate: %s", status.Name, status.Active, status.UpDate.Format("2006-01-02 15:04"))
		if err != nil {
			return err
		}
		fmt.Println(f)
	}
	return err
}
