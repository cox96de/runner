package kube

import (
	"runtime"
	"testing"

	"github.com/cox96de/runner/testtool"

	"github.com/cox96de/runner/engine"
	corev1 "k8s.io/api/core/v1"
)

func Test_newCompiler(t *testing.T) {
	if runtime.GOOS == "windows" {
		// FIXME: All tests should be skipped in windows automatically.
		t.Skip("Skip test on windows")
	}
	c := newCompiler("cox96de/runner", "/executor")
	compileResult := c.Compile("test", &engine.KubeSpec{
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
	testtool.DeepEqualObject(t, compileResult.pod, "testdata/pod2.json")
}
