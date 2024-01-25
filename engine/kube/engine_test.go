package kube

import (
	"context"
	"runtime"
	"testing"

	"github.com/cox96de/runner/entity"

	"github.com/cox96de/runner/testtool"
	"gotest.tools/v3/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

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
	// FIXME: All tests should be skipped in windows automatically.
	if runtime.GOOS == "windows" {
		t.Skip("Skip test on windows")
	}
	namespace := "kube"
	clientset := fake.NewSimpleClientset()
	e, err := NewEngine(clientset, &Option{
		ExecutorImage: "executor_image",
		ExecutorPath:  "/path/executor",
		Namespace:     namespace,
	})
	assert.NilError(t, err)
	t.Run("no_spec", func(t *testing.T) {
		_, err := e.CreateRunner(context.Background(), &entity.Job{
			ID:     1,
			RunsOn: &entity.RunsOn{},
		})
		assert.ErrorContains(t, err, "is nil")
	})
	t.Run("success", func(t *testing.T) {
		r, err := e.CreateRunner(context.Background(), &entity.Job{
			ID: 1,
			RunsOn: &entity.RunsOn{
				Docker: &entity.Docker{
					Containers: []*entity.Container{
						{Image: "debian", Name: "test", VolumeMounts: []*entity.VolumeMount{{Name: "test", MountPath: "/test"}}},
					},
					Volumes: []*entity.Volume{{
						Name:     "test",
						EmptyDir: &entity.EmptyDirVolumeSource{},
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
		// TODO: hard to mock RESTClient
	})
}

func TestNewEngine(t *testing.T) {
	t.Run("port_forward", func(t *testing.T) {
		_, err := NewEngine(nil, &Option{UsePortForward: true})
		assert.ErrorContains(t, err, "kube config is nil, it is required when UsePortForward is true")
	})
}
