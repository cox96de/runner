package kube

import (
	"runtime"
	"testing"

	"github.com/cox96de/runner/api"

	"github.com/cox96de/runner/testtool"
)

func Test_newCompiler(t *testing.T) {
	if runtime.GOOS == "windows" {
		// FIXME: All tests should be skipped in windows automatically.
		t.Skip("Skip test on windows")
	}
	c := newCompiler("cox96de/runner", "/executor")
	compileResult := c.Compile("test", &api.RunsOn{
		Docker: &api.Docker{
			Containers: []*api.Container{
				{Image: "debian", Name: "test", VolumeMounts: []*api.VolumeMount{{Name: "test", MountPath: "/test"}}},
			},
			Volumes: []*api.Volume{{
				Name:     "test",
				EmptyDir: &api.EmptyDirVolumeSource{},
			}},
		},
	})
	testtool.DeepEqualObject(t, compileResult.pod, "testdata/pod2.json")
}
