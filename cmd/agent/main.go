package main

import (
	"context"
	"net/http"

	"github.com/spf13/viper"

	"github.com/cox96de/runner/api/httpserverclient"

	"github.com/cox96de/runner/app/agent"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	command := GetAgentCommand()
	root := &cobra.Command{}
	root.AddCommand(command)
	err := root.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func GetAgentCommand() *cobra.Command {
	vv := viper.New()
	var configFilePath string
	c := &cobra.Command{
		Use: "agent",
		Run: func(cmd *cobra.Command, args []string) {
			if len(configFilePath) > 0 {
				vv.SetConfigFile(configFilePath)
			}
			var config Config
			err := vv.UnmarshalExact(&config)
			if err != nil {
				log.Fatalf("failed to load config: %v", err)
			}
			log.Infof("%+v", vv.AllKeys())
			log.Infof("%+v", vv.AllSettings())
			log.SetLevel(log.DebugLevel)
			err = RunAgent(&config)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
	flags := c.Flags()
	flags.StringVarP(&configFilePath, "config", "c", "", "config file path")

	checkError(bindArg(flags, vv, &arg{
		ArgKey:    "server_url",
		FlagName:  "server_url",
		FlagUsage: "server url",
		Env:       "RUNNER_SERVER_URL",
	}))
	checkError(bindArg(flags, vv, &arg{
		ArgKey:    "label",
		FlagName:  "label",
		FlagUsage: "label",
		Env:       "RUNNER_LABEL",
	}))
	checkError(bindArg(flags, vv, &arg{
		ArgKey:    "engine.name",
		FlagName:  "engine",
		FlagUsage: "engine's type, support: shell, kube, vm",
		Env:       "RUNNER_ENGINE_ENGINE",
	}))
	checkError(bindArg(flags, vv, &arg{
		ArgKey:    "engine.kube.executor_image",
		FlagName:  "engine.kube.executor_image",
		FlagUsage: "the image of executor (kube engine)",
		Env:       "RUNNER_ENGINE_KUBE_EXECUTOR_IMAGE",
	}))
	checkError(bindArg(flags, vv, &arg{
		ArgKey:    "engine.kube.executor_path",
		FlagName:  "engine.kube.executor_path",
		FlagUsage: "the executor binary path in executor image (kube engine)",
		Env:       "RUNNER_ENGINE_KUBE_EXECUTOR_PATH",
	}))
	checkError(bindArg(flags, vv, &arg{
		ArgKey:    "engine.kube.namespace",
		FlagName:  "engine.kube.namespace",
		FlagUsage: "the namespace of executor pod created (kube engine)",
		Env:       "RUNNER_ENGINE_KUBE_NAMESPACE",
	}))
	return c
}

func checkError(err error) {
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func RunAgent(config *Config) error {
	engine, err := ComposeEngine(config)
	if err != nil {
		return errors.WithMessage(err, "failed to compose engine")
	}
	serverClient, err := httpserverclient.NewClient(&http.Client{}, config.ServerURL)
	if err != nil {
		return errors.WithMessage(err, "failed to create server client")
	}
	agent := agent.NewAgent(engine, serverClient, config.Label)
	log.Infof("agent is running on '%s' with label: %s", config.ServerURL, config.Label)
	return agent.Run(context.Background())
}
