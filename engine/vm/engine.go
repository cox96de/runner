package vm

import (
	"context"
	"fmt"

	"github.com/cox96de/runner/util"

	corev1 "k8s.io/api/core/v1"

	"github.com/cox96de/runner/api"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/engine"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Option struct {
	// ExecutorPath is the path of the executor binary in the executor image.
	ExecutorPath string
	// KubeConfig is the kube config used to connect to the kubernetes cluster.
	// It's required when UsePortForward is true.
	KubeConfig *rest.Config
	//// UsePortForward is true, the runner will use port-forward to connect to the executor.
	//// It is useful when the runner is running outside the kubernetes cluster.
	//UsePortForward bool
	// Namespace is the namespace to create the runner pod.
	Namespace string
	// RuntimeImage is the image of contains the vm runtime binary and qemu binary.
	RuntimeImage string
	// VMImageRoot is the root directory in host witch contains qcow2 files.
	// The VMImageRoot will be mounted to the runner pod.
	VMImageRoot string

	Volumes []string
}
type Engine struct {
	client              kubernetes.Interface
	executorPath        string
	namespace           string
	kubeConfig          *rest.Config
	runtimeImage        string
	vmImageRoot         string
	runtimeVolumes      []corev1.Volume
	runtimeVolumeMounts []corev1.VolumeMount
}

func NewEngine(client kubernetes.Interface, opt *Option) (*Engine, error) {
	volumes, volumeMounts, err := parseVolumes(opt.Volumes)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to parse volumes")
	}
	return &Engine{
		client:              client,
		executorPath:        opt.ExecutorPath,
		namespace:           opt.Namespace,
		kubeConfig:          opt.KubeConfig,
		runtimeImage:        opt.RuntimeImage,
		vmImageRoot:         opt.VMImageRoot,
		runtimeVolumeMounts: volumeMounts,
		runtimeVolumes:      volumes,
	}, nil
}

func (e *Engine) Ping(ctx context.Context) error {
	_, err := e.client.CoreV1().Namespaces().Get(ctx, e.namespace, metav1.GetOptions{})
	return err
}

func (e *Engine) CreateRunner(ctx context.Context, logProvider engine.LogProvider, spec *api.Job) (engine.Runner, error) {
	if spec.RunsOn == nil || spec.RunsOn.VM == nil {
		return nil, errors.Errorf("VM spec is required")
	}
	c := newCompiler(&createCompilerOption{
		RuntimeImage:        e.runtimeImage,
		ExecutorPath:        e.executorPath,
		ImageBaseDir:        e.vmImageRoot,
		RuntimeVolumeMounts: e.runtimeVolumeMounts,
		RuntimeVolumes:      e.runtimeVolumes,
		CPU:                 spec.RunsOn.VM.CPU,
		Memory:              spec.RunsOn.VM.Memory,
	})
	compile, err := c.Compile(fmt.Sprintf("vm-%d-%s", spec.Execution.ID, util.RandomLower(5)), spec.RunsOn)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to compile compile vm-runner")
	}
	r := &Runner{
		client:      e.client,
		pod:         compile.pod,
		port:        int32(executorPort),
		namespace:   e.namespace,
		logProvider: logProvider,
	}
	return r, nil
}
