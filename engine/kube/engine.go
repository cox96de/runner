package kube

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/cox96de/runner/api"

	"github.com/cox96de/runner/engine"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Engine struct {
	client         kubernetes.Interface
	executorImage  string
	executorPath   string
	namespace      string
	kubeConfig     *rest.Config
	usePortForward bool
}

type Option struct {
	// ExecutorImage is the image of contains the executor binary.
	ExecutorImage string
	// ExecutorPath is the path of the executor binary in the executor image.
	ExecutorPath string
	// KubeConfig is the kube config used to connect to the kubernetes cluster.
	// It's required when UsePortForward is true.
	KubeConfig *rest.Config
	// UsePortForward is true, the runner will use port-forward to connect to the executor.
	// It is useful when the runner is running outside the kubernetes cluster.
	UsePortForward bool
	// Namespace is the namespace to create the runner pod.
	Namespace string
}

func NewEngine(client kubernetes.Interface, opt *Option) (*Engine, error) {
	if opt.UsePortForward && opt.KubeConfig == nil {
		return nil, errors.New("kube config is nil, it is required when UsePortForward is true")
	}
	return &Engine{
		client:         client,
		executorImage:  opt.ExecutorImage,
		executorPath:   opt.ExecutorPath,
		namespace:      opt.Namespace,
		kubeConfig:     opt.KubeConfig,
		usePortForward: opt.UsePortForward,
	}, nil
}

func (e *Engine) Ping(ctx context.Context) error {
	_, err := e.client.CoreV1().Namespaces().Get(ctx, e.namespace, metav1.GetOptions{})
	return err
}

func (e *Engine) CreateRunner(ctx context.Context, spec *api.Job) (engine.Runner, error) {
	if spec.RunsOn == nil || spec.RunsOn.Docker == nil {
		return nil, errors.New("runs_on.docker is nil")
	}
	c := newCompiler(e.executorImage, e.executorPath)
	compile := c.Compile(strconv.FormatInt(spec.ID, 10), spec.RunsOn)
	r := &Runner{
		defaultContainer: spec.RunsOn.Docker.DefaultContainer,
		client:           e.client,
		pod:              compile.pod,
		executorPortMap:  c.executorPortMap,
		namespace:        e.namespace,
	}
	if e.usePortForward {
		ports := make([]string, 0, len(c.executorPortMap))
		for _, i := range c.executorPortMap {
			ports = append(ports, fmt.Sprintf("0:%d", i))
		}
		r.portForwardStop = make(chan struct{})
		portForwarder, err := newPortForward(e.client, e.kubeConfig, r.namespace, r.pod.Name, ports,
			// TODO: handle stdout output and stderr output, use os.Stdout here
			r.portForwardStop, os.Stdout, os.Stdout)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		r.portForwarder = portForwarder
	}
	return r, nil
}
