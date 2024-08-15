package main

import (
	"github.com/jinzhu/configor"
	"github.com/pkg/errors"
)

type Config struct {
	ServerURL string `json:"server_url" yaml:"server_url"`
	Label     string `json:"label" yaml:"label"`
	Engine    Engine `json:"engine" yaml:"engine"`
}

type Engine struct {
	Name string `json:"name" yaml:"name"`
	Kube Kube   `json:"kube" yaml:"kube"`
}

type Kube struct {
	Config         string `json:"config" yaml:"config"`
	ExecutorImage  string `json:"executor_image" yaml:"executor_image"`
	ExecutorPath   string `json:"executor_path" yaml:"executor_path"`
	Namespace      string `json:"namespace" yaml:"namespace"`
	UsePortForward bool   `json:"use_port_forward" yaml:"use_port_forward"`
}

func LoadConfig(path string) (*Config, error) {
	configLoader := configor.New(&configor.Config{
		Verbose:   true,
		ENVPrefix: "AGENT_CONFIG",
	})
	var config Config
	err := configLoader.Load(&config, path)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to load config")
	}
	return &config, nil
}
