package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/uwine4850/anthill/internal/pathutils"
	"gopkg.in/yaml.v3"
)

type PluginsConfig struct {
	Plugins []string
}

func ParsePlugins(configPath string) (*PluginsConfig, error) {
	if err := pathutils.Exists(configPath); err != nil {
		return nil, err
	}

	f, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var pluginsConfig PluginsConfig
	if err := yaml.Unmarshal(f, &pluginsConfig); err != nil {
		return nil, err
	}
	if err := checkPluginsPath(pluginsConfig.Plugins); err != nil {
		return nil, err
	}
	return &pluginsConfig, nil
}

func checkPluginsPath(workers []string) error {
	for i := 0; i < len(workers); i++ {
		if err := pathutils.Exists(workers[i]); err != nil {
			return err
		}
		isFile, err := pathutils.IsFile(workers[i])
		if err != nil {
			return err
		}
		if !isFile {
			return fmt.Errorf("plugin %s must be a file", workers[i])
		}
		if filepath.Ext(workers[i]) != ".so" {
			return fmt.Errorf("plugin %s must have .so extension", workers[i])
		}
	}
	return nil
}
