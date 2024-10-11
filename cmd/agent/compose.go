package main

import (
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/engine"
	"github.com/cox96de/runner/engine/kube"
	"github.com/cox96de/runner/engine/shell"
	"github.com/cox96de/runner/engine/vm"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func ComposeEngine(c *Config) (engine.Engine, error) {
	switch c.Engine.Name {
	case "shell":
		return shell.NewEngine(), nil
	case "kube":
		return ComposeKubeEngine(&c.Engine.Kube)
	case "vm":
		return ComposeVMEngine(&c.Engine.VM)
	default:
		return nil, errors.Errorf("unsupported engine '%s'", c.Engine.Name)
	}
}

func ComposeKubeEngine(c *Kube) (engine.Engine, error) {
	var (
		clientset *kubernetes.Clientset
		kubeconf  *rest.Config
		err       error
	)
	if c.Config == "" {
		clientset, kubeconf, err = kube.ComposeKubeClientInKube()
	} else {
		clientset, kubeconf, err = kube.ComposeKubeClientFromFile(c.Config)
	}
	if err != nil {
		return nil, errors.WithMessage(err, "failed to compose kube client")
	}
	return kube.NewEngine(clientset, &kube.Option{
		ExecutorImage:  c.ExecutorImage,
		ExecutorPath:   c.ExecutorPath,
		KubeConfig:     kubeconf,
		UsePortForward: c.UsePortForward,
		Namespace:      c.Namespace,
	})
}

func ComposeVMEngine(c *VM) (engine.Engine, error) {
	var (
		clientset *kubernetes.Clientset
		kubeconf  *rest.Config
		err       error
	)
	if c.Config == "" {
		clientset, kubeconf, err = kube.ComposeKubeClientInKube()
	} else {
		clientset, kubeconf, err = kube.ComposeKubeClientFromFile(c.Config)
	}
	if err != nil {
		return nil, errors.WithMessage(err, "failed to compose kube client")
	}
	var volumes []string
	if len(c.Volumes) > 0 {
		volumes = strings.Split(c.Volumes, ":")
	}
	return vm.NewEngine(clientset, &vm.Option{
		ExecutorPath: c.ExecutorPath,
		KubeConfig:   kubeconf,
		Namespace:    c.Namespace,
		RuntimeImage: c.RuntimeImage,
		VMImageRoot:  c.ImageRoot,
		Volumes:      volumes,
	})
}
