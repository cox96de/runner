package main

import "github.com/cox96de/runner/composer"

type Config struct {
	HTTP *HTTP `mapstructure:"http" yaml:"http"`
	GRPC *GRPC `mapstructure:"grpc" yaml:"grpc"`
	// DB is the database configuration.
	DB *composer.DB `mapstructure:"db" yaml:"db"`
	// Locker is the config of distribute locker.
	Locker     *Locker     `mapstructure:"locker" yaml:"locker"`
	LogStorage *LogStorage `mapstructure:"log_storage" yaml:"log_storage"`
	Event      *Event      `mapstructure:"event" yaml:"event"`
}

type HTTP struct {
	Port int `mapstructure:"port" yaml:"port"`
}

type GRPC struct {
	Port int `mapstructure:"port" yaml:"port"`
}

type Locker struct {
	// Dialect is the locker dialect, support redis, db.
	Backend string          `mapstructure:"backend" yaml:"backend"`
	Redis   *composer.Redis `mapstructure:"redis" yaml:"redis"`
}

type LogStorage struct {
	Redis      *composer.Redis `mapstructure:"redis" yaml:"redis"`
	LogArchive *LogArchive     `mapstructure:"archive" yaml:"archive"`
}

type LogArchive struct {
	Backend string `mapstructure:"backend" yaml:"backend"`
	BaseDir string `mapstructure:"base_dir" yaml:"base_dir"`
}

type Event struct {
	HTTPEndPoint string `mapstructure:"http_endpoint" yaml:"http_endpoint"`
}
