package main

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ServerURL string `json:"server_url" yaml:"server_url"`
	Engine    Engine `json:"engine" yaml:"engine"`
}

type Engine struct {
	Name string `json:"name" yaml:"name"`
	Kube Kube   `json:"kube" yaml:"kube"`
}

type Kube struct {
	Config string `json:"config" yaml:"config"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to read config file")
	}
	var config Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		return nil, errors.WithMessage(err, "failed to unmarshal config file")
	}
	return &config, nil
}
