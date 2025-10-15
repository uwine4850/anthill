package config

import (
	"os"

	"github.com/uwine4850/anthill/internal/pathutils"
	"gopkg.in/yaml.v3"
)

type WorkerConfig struct {
	Name   string
	Reload bool
	Type   string
	Args   map[string]string
}

type WorkersConfig struct {
	Workers []WorkerConfig
}

func ParseWorkers(configPath string) (*WorkersConfig, error) {
	if err := pathutils.Exists(configPath); err != nil {
		return nil, err
	}

	f, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var workersConfig WorkersConfig
	if err := yaml.Unmarshal(f, &workersConfig); err != nil {
		return nil, err
	}
	return &workersConfig, nil
}
