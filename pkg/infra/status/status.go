package status

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/uwine4850/anthill/pkg/infra/parsecnf"
	"github.com/uwine4850/anthill/pkg/infra/socket"
)

type Status interface {
	Init(workersConfig *parsecnf.WorkersConfig)
	SetRunning(name string) error
	SetStopped(name string) error
	SetDone(name string) error
	Get() map[string]WorkerStatusData
}

type WorkerStatusData struct {
	Name   string
	Active bool
	UpDate time.Time
	Done   bool
}

type StatusResponse struct {
	WorkerStatus map[string]WorkerStatusData
	Error        string
}

type WorkerStatus struct {
	workerAntsStatus map[string]WorkerStatusData
}

func NewStatus() *WorkerStatus {
	return &WorkerStatus{
		workerAntsStatus: make(map[string]WorkerStatusData),
	}
}

func (s *WorkerStatus) Init(workersConfig *parsecnf.WorkersConfig) {
	for i := 0; i < len(workersConfig.Workers); i++ {
		w := workersConfig.Workers[i]
		s.workerAntsStatus[w.Name] = WorkerStatusData{
			Name: w.Name,
		}
	}
}

func (s *WorkerStatus) SetRunning(name string) error {
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

func (s *WorkerStatus) SetStopped(name string) error {
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

func (s *WorkerStatus) SetDone(name string) error {
	w, ok := s.workerAntsStatus[name]
	if ok {
		w.Active = false
		w.UpDate = time.Time{}
		w.Done = true
		s.workerAntsStatus[name] = w
	} else {
		return fmt.Errorf("worker %s not exists", name)
	}
	return nil
}

func (s *WorkerStatus) Get() map[string]WorkerStatusData {
	return s.workerAntsStatus
}

func SendResponse(conn net.Conn, status Status) error {
	err := socket.SendRequest(conn, &StatusResponse{WorkerStatus: status.Get()})
	if err != nil {
		return err
	}
	return nil
}

func SendWorkerResponse(conn net.Conn, workerName string, status Status) error {
	workerStatus, ok := status.Get()[workerName]
	if !ok {
		return fmt.Errorf("worker %s not exists", workerName)
	}
	err := socket.SendRequest(conn, &StatusResponse{WorkerStatus: map[string]WorkerStatusData{workerName: workerStatus}})
	if err != nil {
		return err
	}
	return nil
}

func CheckAllStatus() error {
	req := socket.Request{Action: "status"}
	conn, err := socket.ConnectToOrchestrator()
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := socket.SendRequest(conn, req); err != nil {
		return err
	}
	var resp StatusResponse
	if err := socket.ReadRequest(conn, &resp); err != nil {
		return err
	}

	for _, status := range resp.WorkerStatus {
		f, err := fmt.Printf("Name: %s | Active: %v, | UpDate: %s", status.Name, status.Active, status.UpDate.Format("2006-01-02 15:04"))
		if err != nil {
			return err
		}
		fmt.Println(f)
	}
	return err
}

func CheckStatus(name string) (*WorkerStatusData, error) {
	conn, err := socket.ConnectToOrchestrator()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req := socket.Request{Action: "status", Name: name}
	if err := socket.SendRequest(conn, &req); err != nil {
		return nil, err
	}

	var resp StatusResponse
	if err := socket.ReadRequest(conn, &resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}
	w := resp.WorkerStatus[name]
	return &w, nil
}
