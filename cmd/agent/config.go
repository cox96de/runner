package main

type Config struct {
	ServerURL   string `mapstructure:"server_url" yaml:"server_url"`
	Concurrency int    `mapstructure:"concurrency" yaml:"concurrency"`
	Label       string `mapstructure:"label" yaml:"label"`
	Engine      Engine `mapstructure:"engine" yaml:"engine"`
}

type Engine struct {
	Name string `mapstructure:"name" yaml:"name"`
	Kube Kube   `mapstructure:"kube" yaml:"kube"`
	VM   VM     `mapstructure:"vm" yaml:"vm"`
}

type Kube struct {
	Config         string `mapstructure:"config" yaml:"config"`
	ExecutorImage  string `mapstructure:"executor_image" yaml:"executor_image"`
	ExecutorPath   string `mapstructure:"executor_path" yaml:"executor_path"`
	Namespace      string `mapstructure:"namespace" yaml:"namespace"`
	UsePortForward bool   `mapstructure:"use_port_forward" yaml:"use_port_forward"`
}
type VM struct {
	Config       string `mapstructure:"config" yaml:"config"`
	RuntimeImage string `mapstructure:"runtime_image" yaml:"executor_image"`
	ExecutorPath string `mapstructure:"executor_path" yaml:"executor_path"`
	Namespace    string `mapstructure:"namespace" yaml:"namespace"`
	Volumes      string `mapstructure:"volumes" yaml:"volumes"`
	ImageRoot    string `mapstructure:"image_root" yaml:"image_root"`
}
