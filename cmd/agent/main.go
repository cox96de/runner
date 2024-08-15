package main

import (
	"context"
	"net/http"

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
	var configPath string
	c := &cobra.Command{
		Use: "agent",
		Run: func(cmd *cobra.Command, args []string) {
			err := RunAgent(configPath)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
	flags := c.Flags()
	flags.StringVarP(&configPath, "config", "c", "config.yaml", "path to config file")
	return c
}

func RunAgent(configfile string) error {
	config, err := LoadConfig(configfile)
	if err != nil {
		return errors.WithMessage(err, "failed to load config")
	}
	log.SetLevel(log.DebugLevel)
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
