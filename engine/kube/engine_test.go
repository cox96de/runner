package kube

import (
	"context"
	"runtime"
	"testing"

	"github.com/cox96de/runner/engine"
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
	err := NewEngine(clientset, &Option{Namespace: namespace}).Ping(context.Background())
	assert.NilError(t, err)
}

func TestEngine_CreateRunner(t *testing.T) {
	// FIXME: All tests should be skipped in windows automatically.
	if runtime.GOOS == "windows" {
		t.Skip("Skip test on windows")
	}
	namespace := "kube"
	clientset := fake.NewSimpleClientset()
	e := NewEngine(clientset, &Option{
		ExecutorImage: "executor_image",
		ExecutorPath:  "/path/executor",
		Namespace:     namespace,
	})
	t.Run("no_spec", func(t *testing.T) {
		_, err := e.CreateRunner(context.Background(), &engine.RunnerSpec{
			ID:   "test",
			Kube: nil,
		})
		assert.ErrorContains(t, err, "kube spec is nil")
	})
	t.Run("no_spec", func(t *testing.T) {
		r, err := e.CreateRunner(context.Background(), &engine.RunnerSpec{
			ID: "test",
			Kube: &engine.KubeSpec{
				Containers: []*engine.Container{
					{Image: "debian", Name: "test", VolumeMounts: []corev1.VolumeMount{{Name: "test", MountPath: "/test"}}},
				},
				Volumes: []corev1.Volume{{
					Name: "test",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				}},
			},
		})
		assert.NilError(t, err)
		runner := r.(*Runner)
		assert.DeepEqual(t, runner.namespace, namespace)
		testtool.DeepEqualObject(t, runner.pod, "testdata/pod.json")
	})
}
