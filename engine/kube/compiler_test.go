package kube

import (
	"testing"

	"github.com/cox96de/runner/engine"
	"gotest.tools/v3/assert"
	corev1 "k8s.io/api/core/v1"
)

func Test_newCompiler(t *testing.T) {
	c := newCompiler("cox96de/runner", "/executor")
	_, err := c.Compile("test", &engine.KubeSpec{
		Containers: []*engine.Container{
			{Image: "debian", Name: "test", VolumeMounts: []corev1.VolumeMount{{Name: "test", MountPath: "/test"}}},
		},
		Volumes: []corev1.Volume{{
			Name: "test",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		}},
	})
	assert.NilError(t, err)
}
