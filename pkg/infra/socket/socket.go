package socket

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/uwine4850/anthill/pkg/config"
)

func ConnectToOrchestrator() (net.Conn, error) {
	conn, err := net.Dial("unix", config.ANTHILL_SOCKET_PATH)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to orchestrator: %s", err)
	}
	return conn, nil
}

func SendRequest(conn net.Conn, req any) error {
	enc := json.NewEncoder(conn)
	err := enc.Encode(req)
	if err != nil {
		return err
	}
	return nil
}

func ReadRequest[T any](conn net.Conn, resp *T) error {
	dec := json.NewDecoder(conn)
	if err := dec.Decode(resp); err != nil {
		return err
	}
	return nil
}
