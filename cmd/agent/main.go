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
	// configor := &Configor{}
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
	// cc = NewConfigure(flags)
	_ = flags.String("server-url", "", "server url")
	checkError(vv.BindPFlag("server_url", flags.Lookup("server-url")))
	checkError(vv.BindEnv("server_url", "RUNNER_SERVER_URL"))

	_ = flags.String("label", "", "label")
	checkError(vv.BindPFlag("label", flags.Lookup("label")))
	checkError(vv.BindEnv("label", "RUNNER_LABEL"))

	_ = flags.String("engine", "", "engine.kube.executor_image")
	checkError(vv.BindPFlag("engine.name", flags.Lookup("engine")))
	checkError(vv.BindEnv("engine.name", "RUNNER_ENGINE_ENGINE"))

	_ = flags.String("engine.kube.executor_image", "", "engine.kube.executor_image")
	checkError(vv.BindPFlag("engine.kube.executor_image", flags.Lookup("engine.kube.executor_image")))
	checkError(vv.BindEnv("engine.kube.executor_image", "RUNNER_ENGINE_KUBE_EXECUTOR_IMAGE"))

	_ = flags.String("engine.kube.namespace", "", "engine.kube.namespace")
	checkError(vv.BindPFlag("engine.kube.namespace", flags.Lookup("engine.kube.namespace")))
	checkError(vv.BindEnv("engine.kube.namespace", "RUNNER_ENGINE_KUBE_NAMESPACE"))

	_ = flags.String("engine.kube.executor_path", "/executor", "engine.kube.executor_path")
	checkError(vv.BindPFlag("engine.kube.executor_path", flags.Lookup("engine.kube.executor_path")))
	checkError(vv.BindEnv("engine.kube.executor_path", "RUNNER_ENGINE_KUBE_EXECUTOR_PATH"))

	// cc.StringVarP(&conf.Engine.Name, "engine", "", "shell", "the name of the engine")
	// cc.StringVarP(&conf.Engine.Kube.Namespace, "engine.kube.namespace", "", "shell", "the namespace of executor for kube engine")
	// cc.StringVarP(&conf.Engine.Kube.ExecutorPath, "engine.kube.executor-path", "", "/executor", "the executor path of executor for kube engine")
	// cc.StringVarP(&conf.Engine.Kube.ExecutorImage, "engine.kube.executor-image", "", "", "the executor image of executor for kube engine")
	// configor.ParseFlag(flags)
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
