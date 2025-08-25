package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type CommandsConfig struct {
	Commands map[string]CommandSpec `yaml:"commands"`
}

type CommandSpec struct {
	Command        string `yaml:"command"`
	Description    string `yaml:"description"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
}

func LoadCommandsConfig(path string) (*CommandsConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config CommandsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}
