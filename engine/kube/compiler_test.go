package kube

import (
	"runtime"
	"testing"

	"github.com/cox96de/runner/entity"

	"github.com/cox96de/runner/testtool"
)

func Test_newCompiler(t *testing.T) {
	if runtime.GOOS == "windows" {
		// FIXME: All tests should be skipped in windows automatically.
		t.Skip("Skip test on windows")
	}
	c := newCompiler("cox96de/runner", "/executor")
	compileResult := c.Compile("test", &entity.RunsOn{
		Docker: &entity.Docker{
			Containers: []*entity.Container{
				{Image: "debian", Name: "test", VolumeMounts: []*entity.VolumeMount{{Name: "test", MountPath: "/test"}}},
			},
			Volumes: []*entity.Volume{{
				Name:     "test",
				EmptyDir: &entity.EmptyDirVolumeSource{},
			}},
		},
	})
	testtool.DeepEqualObject(t, compileResult.pod, "testdata/pod2.json")
}
