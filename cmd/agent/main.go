package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cox96de/runner/util"
	"github.com/spf13/viper"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/app/agent"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	command := GetAgentCommand()
	err := command.Execute()
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
			log.SetLevel(log.DebugLevel)
			err = RunAgent(&config)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
	flags := c.Flags()
	flags.StringVarP(&configFilePath, "config", "c", "", "config file path")
	checkError(util.BindStringArg(flags, vv, &util.StringArg{
		ArgKey:    "server_url",
		FlagName:  "server_url",
		FlagUsage: "the server url, such as http://127.0.0.1:8080",
		Env:       "RUNNER_SERVER_URL",
	}))
	checkError(util.BindIntArg(flags, vv, &util.IntArg{
		ArgKey:    "concurrency",
		FlagName:  "concurrency",
		FlagValue: 1,
		FlagUsage: "the max concurrency of running jobs",
		Env:       "RUNNER_CONCURRENCY",
	}))
	checkError(util.BindStringArg(flags, vv, &util.StringArg{
		ArgKey:    "label",
		FlagName:  "label",
		FlagUsage: "the label of agent",
		Env:       "RUNNER_LABEL",
	}))
	checkError(util.BindStringArg(flags, vv, &util.StringArg{
		ArgKey:    "engine.name",
		FlagName:  "engine",
		FlagUsage: "engine's type, support: shell, kube, vm",
		Env:       "RUNNER_ENGINE_ENGINE",
	}))
	checkError(util.BindStringArg(flags, vv, &util.StringArg{
		ArgKey:    "engine.kube.executor_image",
		FlagName:  "engine.kube.executor_image",
		FlagUsage: "the image of executor (kube engine)",
		Env:       "RUNNER_ENGINE_KUBE_EXECUTOR_IMAGE",
	}))
	checkError(util.BindStringArg(flags, vv, &util.StringArg{
		ArgKey:    "engine.kube.executor_path",
		FlagName:  "engine.kube.executor_path",
		FlagUsage: "the executor binary path in executor image (kube engine)",
		Env:       "RUNNER_ENGINE_KUBE_EXECUTOR_PATH",
	}))
	checkError(util.BindStringArg(flags, vv, &util.StringArg{
		ArgKey:    "engine.kube.namespace",
		FlagName:  "engine.kube.namespace",
		FlagUsage: "the namespace of executor pod created (kube engine)",
		Env:       "RUNNER_ENGINE_KUBE_NAMESPACE",
	}))

	checkError(util.BindStringArg(flags, vv, &util.StringArg{
		ArgKey:    "engine.vm.runtime_image",
		FlagName:  "engine.vm.runtime_image",
		FlagUsage: "the image of executor for vm engine",
		Env:       "RUNNER_ENGINE_VM_EXECUTOR_IMAGE",
	}))
	checkError(util.BindStringArg(flags, vv, &util.StringArg{
		ArgKey:    "engine.vm.executor_path",
		FlagName:  "engine.vm.executor_path",
		FlagUsage: "the executor binary path in runtime image (vm engine)",
		Env:       "RUNNER_ENGINE_VM_EXECUTOR_PATH",
	}))
	checkError(util.BindStringArg(flags, vv, &util.StringArg{
		ArgKey:    "engine.vm.namespace",
		FlagName:  "engine.vm.namespace",
		FlagUsage: "the namespace of executor pod created for (vm engine)",
		Env:       "RUNNER_ENGINE_VM_NAMESPACE",
	}))
	checkError(util.BindStringArg(flags, vv, &util.StringArg{
		ArgKey:    "engine.vm.volumes",
		FlagName:  "engine.vm.volumes",
		FlagUsage: "the volume and mounts for each executor pod (vm engine)",
		Env:       "RUNNER_ENGINE_VM_VOLUMES",
	}))
	checkError(util.BindStringArg(flags, vv, &util.StringArg{
		ArgKey:    "engine.vm.image_root",
		FlagName:  "engine.vm.image_root",
		FlagUsage: "the image root path in executor pod (vm engine)",
		Env:       "RUNNER_ENGINE_VM_IMAGE_ROOT",
	}))
	return c
}

func checkError(err error) {
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func RunAgent(config *Config) error {
	log.SetReportCaller(true)
	engine, err := ComposeEngine(config)
	if err != nil {
		return errors.WithMessage(err, "failed to compose engine")
	}
	serverClient, err := ComposeRunnerClient(config.ServerURL)
	if err != nil {
		return errors.WithMessage(err, "failed to create server client")
	}
	agent := agent.NewAgent(engine, serverClient, config.Label)
	log.Infof("agent is running on '%s' with label: %s", config.ServerURL, config.Label)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	ctx, forceCancel := context.WithCancel(context.Background())
	go func() {
		<-sigChan
		agent.GracefulShutdown()
		log.Infof("await running job to be completed. Try again to force exit")
		<-sigChan
		forceCancel()
	}()

	return agent.Run(ctx, config.Concurrency, time.Second*10)
}
