package main

import (
	"net/http"
	"net/url"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/api/httpserverclient"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

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

func ComposeRunnerClient(u string) (api.ServerClient, error) {
	parse, err := url.Parse(u)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to parse runner url")
	}
	switch parse.Scheme {
	case "http", "https":
		return httpserverclient.NewClient(&http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		}, u)
	case "grpc":
		conn, err := grpc.NewClient(parse.Host, grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithStatsHandler(otelgrpc.NewClientHandler()))
		if err != nil {
			return nil, err
		}
		return api.NewServerClient(conn), nil
	default:
		return nil, errors.Errorf("unsupported scheme '%s'", parse.Scheme)
	}
}
