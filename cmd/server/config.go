package main

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Config struct {
	// Port is the port to listen on.
	Port int `json:"port" yaml:"port"`
	// DB is the database configuration.
	DB *DB `json:"db" yaml:"db"`
}

type DB struct {
	// Dialect is the database dialect, support sqlite, mysql, postgres.
	Dialect string `json:"dialect" yaml:"dialect"`
	DSN     string `json:"dsn" yaml:"dsn"`
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
