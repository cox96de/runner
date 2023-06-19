package kube

import (
	"context"

	"github.com/cox96de/runner/engine"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Engine struct {
	client        kubernetes.Interface
	executorImage string
	executorPath  string
	namespace     string
}

type Option struct {
	ExecutorImage string
	ExecutorPath  string
	Namespace     string
}

func NewEngine(client kubernetes.Interface, opt *Option) *Engine {
	return &Engine{
		client: client, executorImage: opt.ExecutorImage, executorPath: opt.ExecutorPath,
		namespace: opt.Namespace,
	}
}

func (e *Engine) Ping(ctx context.Context) error {
	_, err := e.client.CoreV1().Namespaces().Get(ctx, e.namespace, metav1.GetOptions{})
	return err
}

func (e *Engine) CreateRunner(ctx context.Context, spec *engine.RunnerSpec) (engine.Runner, error) {
	return nil, errors.New("not implemented")
}
