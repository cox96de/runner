package main

import (
	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/engine"
	"github.com/cox96de/runner/engine/kube"
	"github.com/cox96de/runner/engine/shell"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func ComposeEngine(c *Config) (engine.Engine, error) {
	switch c.Engine.Name {
	case "shell":
		return shell.NewEngine(), nil
	case "kube":
		return ComposeKubeEngine(&c.Engine.Kube)
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
