package main

import (
	"github.com/cox96de/runner/composer"
)

type Config struct {
	GithubAppID int64  `yaml:"github_app_id"`
	PrivateKey  string `yaml:"private_key"`
	// DB is the database configuration.
	DB *composer.DB `yaml:"db"`
	// ExportURL is the URL of the log page.
	ExportURL string `yaml:"export_url"`
	// ListenAddr is the listen url such :80.
	ListenAddr string `yaml:"listen_addr"`
	// BaseURL is the prefix path of the web server.
	BaseURL      string        `yaml:"base_url"`
	CloneStep    []string      `yaml:"clone_step"`
	RunnerServer *RunnerServer `yaml:"runner_server"`
}

type RunnerServer struct {
	DB                 *composer.DB    `yaml:"db"`
	GRPCPort           int             `yaml:"grpc_port"`
	Redis              *composer.Redis `yaml:"redis"`
	LogArchiveDir      string          `yaml:"log_archive_dir"`
	LogArchiveS3       *composer.S3    `yaml:"log_archive_s3"`
	LogArchiveS3Bucket string          `yaml:"log_archive_s3_bucket"`
}
