package vm

import (
	"context"
	"runtime"
	"testing"

	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/testtool"
	"gotest.tools/v3/assert"
)

func TestEngine_CreateRunner(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skip test on Windows")
	}
	volumes, mounts, err := parseVolumes([]string{"type=hostPath,name=mnt,mountPath=/mnt,path=/root/mnt"})
	assert.NilError(t, err)
	namespace := "default"
	e := &Engine{
		executorPath:        "/runner/executor",
		namespace:           namespace,
		kubeConfig:          nil,
		runtimeImage:        "registry.houzhiqiang.cn/runner-vm-runtime:latest",
		vmImageRoot:         "/mnt",
		runtimeVolumes:      volumes,
		runtimeVolumeMounts: mounts,
	}
	runner, err := e.CreateRunner(context.Background(), nil, &api.Job{
		RunsOn: &api.RunsOn{
			Label:  "",
			Docker: nil,
			VM: &api.VM{
				Image:    "debian-11.qcow2",
				CPU:      2,
				MemoryMB: 1024,
			},
		},
		Execution: &api.JobExecution{ID: 1},
	})
	assert.NilError(t, err)
	r := runner.(*Runner)
	testtool.DeepEqualObject(t, r.pod, "testdata/pod.json")
}

func TestNewEngine(t *testing.T) {
	_, err := NewEngine(nil, &Option{
		ExecutorPath: "",
		KubeConfig:   nil,
		Namespace:    "",
		RuntimeImage: "",
		VMImageRoot:  "",
		Volumes:      []string{"type=hostPath,name=mnt,mountPath=/mnt,path=/root/mnt", "type=nfs,name=nfs,mountPath=/mnt,server=192.168.33.1,path=/nfs"},
	})
	assert.NilError(t, err)
}
