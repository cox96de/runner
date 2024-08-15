package main

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	ServerURL string `mapstructure:"server_url" yaml:"server_url"`
	Label     string `mapstructure:"label" yaml:"label"`
	Engine    Engine `mapstructure:"engine" yaml:"engine"`
}

type Engine struct {
	Name string `mapstructure:"name" yaml:"name"`
	Kube Kube   `mapstructure:"kube" yaml:"kube"`
}

type Kube struct {
	Config         string `mapstructure:"config" yaml:"config"`
	ExecutorImage  string `mapstructure:"executor_image" yaml:"executor_image"`
	ExecutorPath   string `mapstructure:"executor_path" yaml:"executor_path"`
	Namespace      string `mapstructure:"namespace" yaml:"namespace"`
	UsePortForward bool   `mapstructure:"use_port_forward" yaml:"use_port_forward"`
}

type arg struct {
	ArgKey    string
	FlagName  string
	FlagValue string
	FlagUsage string
	Env       string
}

func bindArg(flags *pflag.FlagSet, viper *viper.Viper, a *arg) error {
	_ = flags.String(a.FlagName, a.FlagValue, a.FlagUsage)
	err := viper.BindPFlag(a.ArgKey, flags.Lookup(a.FlagName))
	if err != nil {
		return err
	}
	err = viper.BindEnv(a.ArgKey, a.Env)
	if err != nil {
		return err
	}
	return nil
}
