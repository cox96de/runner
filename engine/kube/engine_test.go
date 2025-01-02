package kube

import (
	"context"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/cox96de/runner/util"

	"github.com/cox96de/runner/engine/mock"

	"github.com/cox96de/runner/app/executor/executorpb"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/cox96de/runner/api"

	"github.com/cox96de/runner/testtool"
	"gotest.tools/v3/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var (
	clientset *kubernetes.Clientset
	restconf  *rest.Config
)

func TestMain(m *testing.M) {
	configfile := os.ExpandEnv("$HOME/.kube/config")
	if _, err := os.Lstat(configfile); err == nil {
		clientset, restconf, _ = ComposeKubeClientFromFile(configfile)
	}
	os.Exit(m.Run())
}

func TestEngine_Ping(t *testing.T) {
	// FIXME: All tests should be skipped in windows automatically.
	if runtime.GOOS == "windows" {
		t.Skip("Skip test on windows")
	}
	namespace := "kube"
	clientset := fake.NewSimpleClientset(&corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind: "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
		Spec:   corev1.NamespaceSpec{},
		Status: corev1.NamespaceStatus{},
	})
	e, err := NewEngine(clientset, &Option{Namespace: namespace})
	assert.NilError(t, err)
	err = e.Ping(context.Background())
	assert.NilError(t, err)
}

func TestEngine_CreateRunner(t *testing.T) {
	old := util.RandomString
	util.RandomLower = func(n int) string {
		return "mock-random-string"
	}
	t.Cleanup(func() {
		util.RandomLower = old
	})
	// FIXME: All tests should be skipped in windows automatically.
	if runtime.GOOS == "windows" {
		t.Skip("Skip test on windows")
	}
	namespace := "kube"
	fakeclientset := fake.NewSimpleClientset()
	e, err := NewEngine(fakeclientset, &Option{
		ExecutorImage: "executor_image",
		ExecutorPath:  "/path/executor",
		Namespace:     namespace,
	})
	assert.NilError(t, err)
	t.Run("no_spec", func(t *testing.T) {
		_, err := e.CreateRunner(context.Background(), mock.NewNopLogProvider(), &api.Job{
			ID:     1,
			RunsOn: &api.RunsOn{},
		})
		assert.ErrorContains(t, err, "is nil")
	})
	t.Run("success", func(t *testing.T) {
		r, err := e.CreateRunner(context.Background(), mock.NewNopLogProvider(), &api.Job{
			ID:        1,
			Execution: &api.JobExecution{ID: 1},
			RunsOn: &api.RunsOn{
				Docker: &api.Docker{
					Containers: []*api.Container{
						{Image: "debian", Name: "test", VolumeMounts: []*api.VolumeMount{{Name: "test", MountPath: "/test"}}},
					},
					Volumes: []*api.Volume{{
						Name:     "test",
						EmptyDir: &api.EmptyDirVolumeSource{},
					}},
				},
			},
		})
		assert.NilError(t, err)
		runner := r.(*Runner)
		assert.DeepEqual(t, runner.namespace, namespace)
		testtool.DeepEqualObject(t, runner.pod, "testdata/pod.json")
	})
	t.Run("port_forward", func(t *testing.T) {
		t.Skip("skip test") // TODO: recover it.
		if clientset == nil {
			t.Skip("no client set")
		}
		namespace := "default"
		e, err := NewEngine(clientset, &Option{
			ExecutorImage:  "cox96de/runner-executor:master",
			ExecutorPath:   "/executor",
			Namespace:      namespace,
			UsePortForward: true,
			KubeConfig:     restconf,
		})
		assert.NilError(t, err)
		runner, err := e.CreateRunner(context.Background(), mock.NewNopLogProvider(), &api.Job{
			Name: "test",
			RunsOn: &api.RunsOn{
				Docker: &api.Docker{
					DefaultContainer: "test",
					Containers: []*api.Container{
						{Image: "debian", Name: "test", VolumeMounts: []*api.VolumeMount{{Name: "test", MountPath: "/test"}}},
					},
					Volumes: []*api.Volume{{
						Name:     "test",
						EmptyDir: &api.EmptyDirVolumeSource{},
					}},
				},
			},
		})
		assert.NilError(t, err)
		t.Cleanup(func() {
			_ = runner.Stop(context.Background())
		})
		err = runner.Start(context.Background())
		assert.NilError(t, err)
		executor, err := runner.GetExecutor(context.Background())
		assert.NilError(t, err)
		environment, err := executor.Environment(context.Background(), &executorpb.EnvironmentRequest{})
		assert.NilError(t, err)
		assert.Assert(t, strings.Contains(strings.Join(environment.Environment, ""),
			"KUBERNETES_PORT_443_TCP_ADDR"))
	})
}

func TestNewEngine(t *testing.T) {
	t.Run("port_forward", func(t *testing.T) {
		_, err := NewEngine(nil, &Option{UsePortForward: true})
		assert.ErrorContains(t, err, "kube config is nil, it is required when UsePortForward is true")
	})
}
