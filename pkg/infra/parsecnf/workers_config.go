package parsecnf

import (
	"fmt"
	"os"
	"slices"

	"github.com/uwine4850/anthill/internal/pathutils"
	"github.com/uwine4850/anthill/pkg/domain"
	"gopkg.in/yaml.v3"
)

type WorkersConfig struct {
	Workers []domain.WorkerConfig
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
	if err := validateNames(&workersConfig); err != nil {
		return nil, err
	}
	if err := validateAfterList(&workersConfig); err != nil {
		return nil, err
	}
	return &workersConfig, nil
}

func validateNames(workersConfig *WorkersConfig) error {
	var workerName string
	for i := 0; i < len(workersConfig.Workers); i++ {
		name := workersConfig.Workers[i].Name
		if workerName == name {
			return fmt.Errorf("worker <%s> already exists", name)
		} else {
			workerName = name
		}
	}
	return nil
}

func validateAfterList(workersConfig *WorkersConfig) error {
	workersNames := make([]string, len(workersConfig.Workers))
	for i := 0; i < len(workersConfig.Workers); i++ {
		workersNames[i] = workersConfig.Workers[i].Name
	}
	for i := 0; i < len(workersConfig.Workers); i++ {
		for j := 0; j < len(workersConfig.Workers[i].After); j++ {
			after := workersConfig.Workers[i].After[j]
			if !slices.Contains(workersNames, after) {
				return fmt.Errorf("the after field of the worker <%s> contains a non-existent worker <%s>",
					workersConfig.Workers[i].Name, after)
			}
		}
	}
	return nil
}
