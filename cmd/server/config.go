package main

import (
	"time"
)

type Config struct {
	// Port is the port to listen on.
	Port int `mapstructure:"port" yaml:"port"`
	// DB is the database configuration.
	DB *DB `mapstructure:"db" yaml:"db"`
	// Locker is the config of distribute locker.
	Locker     *Locker     `mapstructure:"locker" yaml:"locker"`
	LogStorage *LogStorage `mapstructure:"log_storage" yaml:"log_storage"`
}

type DB struct {
	// Dialect is the database dialect, support sqlite, mysql, postgres.
	Dialect     string `mapstructure:"dialect" yaml:"dialect"`
	DSN         string `mapstructure:"dsn" yaml:"dsn"`
	TablePrefix string `mapstructure:"table_prefix" yaml:"table_prefix"`
}

type Locker struct {
	// Dialect is the locker dialect, support redis, db.
	Backend string `mapstructure:"backend" yaml:"backend"`
	Redis   *Redis `mapstructure:"redis" yaml:"redis"`
}

type Redis struct {
	Addr               string        `mapstructure:"addr" yaml:"addr"`
	Username           string        `mapstructure:"username" yaml:"username"`
	Password           string        `mapstructure:"password" yaml:"password"`
	DB                 int           `mapstructure:"db" yaml:"db"`
	MaxRetries         int           `mapstructure:"max_retries" yaml:"max_retries"`
	MinRetryBackoff    time.Duration `mapstructure:"min_retry_backoff" yaml:"min_retry_backoff"`
	MaxRetryBackoff    time.Duration `mapstructure:"max_retry_backoff" yaml:"max_retry_backoff"`
	DialTimeout        time.Duration `mapstructure:"dial_timeout" yaml:"dial_timeout"`
	ReadTimeout        time.Duration `mapstructure:"read_timeout" yaml:"read_timeout"`
	WriteTimeout       time.Duration `mapstructure:"write_timeout" yaml:"write_timeout"`
	PoolFIFO           bool          `mapstructure:"pool_fifo" yaml:"pool_fifo"`
	PoolSize           int           `mapstructure:"pool_size" yaml:"pool_size"`
	MinIdleConns       int           `mapstructure:"min_idle_conns" yaml:"min_idle_conns"`
	MaxConnAge         time.Duration `mapstructure:"max_conn_age" yaml:"max_conn_age"`
	PoolTimeout        time.Duration `mapstructure:"pool_timeout" yaml:"pool_timeout"`
	IdleTimeout        time.Duration `mapstructure:"idle_timeout" yaml:"idle_timeout"`
	IdleCheckFrequency time.Duration `mapstructure:"idle_check_frequency" yaml:"idle_check_frequency"`
}

type LogStorage struct {
	Redis      *Redis      `mapstructure:"redis" yaml:"redis"`
	LogArchive *LogArchive `mapstructure:"archive" yaml:"archive"`
}

type LogArchive struct {
	Backend string `mapstructure:"backend" yaml:"backend"`
	BaseDir string `mapstructure:"base_dir" yaml:"base_dir"`
}
