package main

import "time"

type Config struct {
	GithubAppID int64  `yaml:"github_app_id"`
	PrivateKey  string `yaml:"private_key"`
	// DB is the database configuration.
	DB *DB `yaml:"db"`
	// ExportURL is the URL of the log page.
	ExportURL string `yaml:"export_url"`
	// ListenAddr is the listen url such :80.
	ListenAddr string `yaml:"listen_addr"`
	// BaseURL is the prefix path of the web server.
	BaseURL      string        `yaml:"base_url"`
	CloneStep    []string      `yaml:"clone_step"`
	RunnerServer *RunnerServer `yaml:"runner_server"`
}

type DB struct {
	// Dialect is the database dialect, support sqlite, mysql, postgres.
	Dialect     string `mapstructure:"dialect" yaml:"dialect"`
	DSN         string `mapstructure:"dsn" yaml:"dsn"`
	TablePrefix string `mapstructure:"table_prefix" yaml:"table_prefix"`
}

type RunnerServer struct {
	DB            *DB    `yaml:"db"`
	GRPCPort      int    `yaml:"grpc_port"`
	Redis         *Redis `yaml:"redis"`
	LogArchiveDir string `yaml:"log_archive_dir"`
}

type Redis struct {
	Addr            string        `mapstructure:"addr" yaml:"addr"`
	Username        string        `mapstructure:"username" yaml:"username"`
	Password        string        `mapstructure:"password" yaml:"password"`
	DB              int           `mapstructure:"db" yaml:"db"`
	MaxRetries      int           `mapstructure:"max_retries" yaml:"max_retries"`
	MinRetryBackoff time.Duration `mapstructure:"min_retry_backoff" yaml:"min_retry_backoff"`
	MaxRetryBackoff time.Duration `mapstructure:"max_retry_backoff" yaml:"max_retry_backoff"`
	DialTimeout     time.Duration `mapstructure:"dial_timeout" yaml:"dial_timeout"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout" yaml:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout" yaml:"write_timeout"`
	PoolFIFO        bool          `mapstructure:"pool_fifo" yaml:"pool_fifo"`
	PoolSize        int           `mapstructure:"pool_size" yaml:"pool_size"`
	PoolTimeout     time.Duration `mapstructure:"pool_timeout" yaml:"pool_timeout"`
	MinIdleConns    int           `mapstructure:"min_idle_conns" yaml:"min_idle_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" yaml:"max_idle_conns"`
	MaxActiveConns  int           `mapstructure:"max_active_conns" yaml:"max_active_conns"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time" yaml:"conn_max_idle_time"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" yaml:"conn_max_lifetime"`
}
