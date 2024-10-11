package main

type Config struct {
	GithubAppID int64  `yaml:"github_app_id"`
	PrivateKey  string `yaml:"private_key"`
	// DB is the database configuration.
	DB *DB `yaml:"db"`
	// ExportURL is the URL of the log page.
	ExportURL string `yaml:"export_url"`
	// RunnerURL is the URL of the runner.
	RunnerURL string `yaml:"runner_url"`
	// ListenAddr is the listen url such :80.
	ListenAddr string `yaml:"listen_addr"`
	// BaseURL is the prefix path of the web server.
	BaseURL   string   `yaml:"base_url"`
	CloneStep []string `yaml:"clone_step"`
}

type DB struct {
	// Dialect is the database dialect, support sqlite, mysql, postgres.
	Dialect     string `mapstructure:"dialect" yaml:"dialect"`
	DSN         string `mapstructure:"dsn" yaml:"dsn"`
	TablePrefix string `mapstructure:"table_prefix" yaml:"table_prefix"`
}
