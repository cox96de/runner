package main

import (
	"context"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cox96de/runner/util"

	"github.com/cox96de/runner/engine"
	"github.com/cox96de/runner/engine/kube"
	"github.com/cox96de/runner/example/dsl"
	"github.com/cox96de/runner/internal/executor"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	flagSet := pflag.NewFlagSet("runner", pflag.ContinueOnError)
	c := &Config{}
	flagSet.StringVar(&c.Engine, "engine", "kube", "The engine to use")
	flagSet.StringVar(&c.Kube.Config, "kube-config", "", "The kube config file")
	flagSet.BoolVar(&c.Kube.PortForwarding, "kube-use-port-forward", false,
		"Use port forward to connect worker pod")
	flagSet.StringVar(&c.Kube.ExecutorImage, "kube-executor-image", "cox96de/runner", "The image of the executor")
	flagSet.StringVar(&c.Kube.ExecutorPath, "kube-executor-path", "/executor", "The path of the executor")
	flagSet.StringVar(&c.Kube.Namespace, "kube-namespace", "default", "The namespace to use")
	err := flagSet.Parse(os.Args[1:])
	checkErr(err)
	e, err := composeEngine(c)
	checkErr(err)
	ctx := context.Background()
	for idx, job := range getJobs() {
		err = runJob(ctx, e, job)
		if err != nil {
			log.Errorf("failed to run job %d: %v", idx, err)
		}

	}
}

func getJobs() []*dsl.Job {
	return []*dsl.Job{
		{
			Runner: &dsl.Runner{Kube: &engine.KubeSpec{
				Containers: []*engine.Container{{
					Name:  "test",
					Image: "debian",
				}},
			}},
			DefaultContainerName: "test",
			Steps:                []*dsl.Step{{Commands: []string{"echo hello"}}},
		},
		{
			Runner: &dsl.Runner{Kube: &engine.KubeSpec{
				Containers: []*engine.Container{{
					Name:  "test",
					Image: "golang:1.20",
				}},
			}},
			DefaultContainerName: "test",
			Steps:                []*dsl.Step{{Commands: []string{"go env -w GOPROXY=https://goproxy.cn,direct", "go install github.com/go-delve/delve/cmd/dlv@latest"}}},
		},
	}
}

func runJob(ctx context.Context, e engine.Engine, job *dsl.Job) error {
	spec := covertDSLToEngineSpec(strings.ToLower(util.RandomString(5)), job)
	runner, err := e.CreateRunner(ctx, spec)
	if err != nil {
		return errors.WithStack(err)
	}
	err = runner.Start(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		err := runner.Stop(ctx)
		if err != nil {
			log.Errorf("failed to stop runner %q: %v", spec.ID, err)
		}
	}()
	for idx, step := range job.Steps {
		containerName := step.ContainerName
		if containerName == "" {
			containerName = job.DefaultContainerName
		}
		exe, err := runner.GetExecutor(ctx, containerName)
		if err != nil {
			return errors.WithMessagef(err, "failed to get executor for step %d", idx)
		}
		err = engine.WaitExecutorReady(ctx, exe, time.Second, time.Second*10)
		if err != nil {
			return errors.WithMessagef(err, "failed to wait executor ready for step %d", idx)
		}
		err = exe.Ping(ctx)
		if err != nil {
			return errors.WithMessagef(err, "failed to ping executor for step %d", idx)
		}
		stepID := strconv.Itoa(idx)
		commands := util.CompileUnixScript(step.Commands)
		err = exe.StartCommand(ctx, stepID, &executor.StartCommandRequest{
			Dir:     step.Workdir,
			Command: []string{"/bin/sh", "-c", "printf '%s' \"$COMMANDS\" | /bin/sh"},
			Env:     map[string]string{"COMMANDS": commands},
		})
		if err != nil {
			return errors.WithMessagef(err, "failed to start command for step %d", idx)
		}
		logReader := exe.GetCommandLogs(ctx, stepID)
		_, err = io.Copy(os.Stdout, logReader)
		if err != nil {
			log.Errorf("failed to copy logs for step %d: %v", idx, err)
		}
		status, err := exe.GetCommandStatus(ctx, stepID)
		if err != nil {
			return errors.WithMessagef(err, "failed to get command status for step %d", idx)
		}
		log.Infof("step %d finished with status %d", idx, status.ExitCode)
	}
	return nil
}

func composeEngine(c *Config) (engine.Engine, error) {
	switch c.Engine {
	case "kube":
		var (
			clientset *kubernetes.Clientset
			config    *rest.Config
			err       error
		)
		if c.Kube.Config != "" {
			clientset, config, err = kube.ComposeKubeClientFromFile(c.Kube.Config)
		} else {
			clientset, config, err = kube.ComposeKubeClientInKube()
		}
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return kube.NewEngine(clientset, &kube.Option{
			ExecutorImage:  c.Kube.ExecutorImage,
			ExecutorPath:   c.Kube.ExecutorPath,
			Namespace:      c.Kube.Namespace,
			KubeConfig:     config,
			UsePortForward: c.Kube.PortForwarding,
		})
	}
	return nil, errors.Errorf("unknown engine %q", c.Engine)
}

func covertDSLToEngineSpec(id string, job *dsl.Job) *engine.RunnerSpec {
	result := &engine.RunnerSpec{
		ID: id,
	}
	kubeDSL := job.Runner.Kube
	if kubeDSL != nil {
		result.Kube = &engine.KubeSpec{
			Containers: kubeDSL.Containers,
			Volumes:    kubeDSL.Volumes,
		}
	}
	return result
}

func checkErr(err error) {
	if err == nil {
		return
	}
	panic(err)
}
