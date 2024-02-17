package main

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Config struct {
	// Port is the port to listen on.
	Port int `json:"port" yaml:"port"`
	// DB is the database configuration.
	DB *DB `json:"db" yaml:"db"`
	// Locker is the config of distribute locker.
	Locker *Locker `json:"locker" yaml:"locker"`
}

type DB struct {
	// Dialect is the database dialect, support sqlite, mysql, postgres.
	Dialect string `json:"dialect" yaml:"dialect"`
	DSN     string `json:"dsn" yaml:"dsn"`
}

type Locker struct {
	// Dialect is the locker dialect, support redis, db.
	Backend string `json:"backend" yaml:"backend"`
	Redis   *Redis `json:"redis" yaml:"redis"`
}

type Redis struct {
	Addr               string        `json:"addr" yaml:"addr"`
	Username           string        `json:"username" yaml:"username"`
	Password           string        `json:"password" yaml:"password"`
	DB                 int           `json:"db" yaml:"db"`
	MaxRetries         int           `json:"max_retries" yaml:"max_retries"`
	MinRetryBackoff    time.Duration `json:"min_retry_backoff" yaml:"min_retry_backoff"`
	MaxRetryBackoff    time.Duration `json:"max_retry_backoff" yaml:"max_retry_backoff"`
	DialTimeout        time.Duration `json:"dial_timeout" yaml:"dial_timeout"`
	ReadTimeout        time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout       time.Duration `json:"write_timeout" yaml:"write_timeout"`
	PoolFIFO           bool          `json:"pool_fifo" yaml:"pool_fifo"`
	PoolSize           int           `json:"pool_size" yaml:"pool_size"`
	MinIdleConns       int           `json:"min_idle_conns" yaml:"min_idle_conns"`
	MaxConnAge         time.Duration `json:"max_conn_age" yaml:"max_conn_age"`
	PoolTimeout        time.Duration `json:"pool_timeout" yaml:"pool_timeout"`
	IdleTimeout        time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
	IdleCheckFrequency time.Duration `json:"idle_check_frequency" yaml:"idle_check_frequency"`
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
