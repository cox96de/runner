package vm

import (
	"context"

	"github.com/cox96de/runner/api"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/engine"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

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
	// RuntimeImage is the image of contains the vm runtime binary and qemu binary.
	RuntimeImage string
	// VMImageRoot is the root directory in host witch contains qcow2 files.
	// The VMImageRoot will be mounted to the runner pod.
	VMImageRoot string
}
type Engine struct {
	client         kubernetes.Interface
	executorImage  string
	executorPath   string
	namespace      string
	kubeConfig     *rest.Config
	usePortForward bool
	runtimeImage   string
	vmImageRoot    string
}

func NewEngine(client kubernetes.Interface, opt *Option) (*Engine, error) {
	return &Engine{
		client:         client,
		executorImage:  opt.ExecutorImage,
		executorPath:   opt.ExecutorPath,
		namespace:      opt.Namespace,
		kubeConfig:     opt.KubeConfig,
		usePortForward: opt.UsePortForward,
		runtimeImage:   opt.RuntimeImage,
		vmImageRoot:    opt.VMImageRoot,
	}, nil
}

func (e *Engine) Ping(ctx context.Context) error {
	_, err := e.client.CoreV1().Namespaces().Get(ctx, e.namespace, metav1.GetOptions{})
	return err
}

func (e *Engine) CreateRunner(ctx context.Context, option *api.Job) (engine.Runner, error) {
	return nil, errors.New("not implemented")
}
