package vm

import (
	"fmt"
	"path/filepath"

	"github.com/cox96de/runner/api"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type compiler struct {
	runtimeImage  string
	executorImage string
	executorPath  string
}

func newCompiler(executorImage string, executorPath string, runtimeImage string) *compiler {
	c := &compiler{
		executorImage: executorImage, executorPath: executorPath, runtimeImage: runtimeImage,
	}
	return c
}

type compileResult struct {
	pod *corev1.Pod
}

func (c *compiler) Compile(id string, spec *api.RunsOn) *compileResult {
	runtimeContainer := c.compileRuntimeContainer(spec.VM)
	result := &compileResult{
		pod: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: id,
			},
			Spec: corev1.PodSpec{
				InitContainers: []corev1.Container{c.compileInitContainer()},
				Containers:     []corev1.Container{runtimeContainer},
				Volumes:        c.compileVolumes(),
			},
		},
	}
	return result
}

func (c *compiler) compileRuntimeContainer(vm *api.VM) corev1.Container {
	return corev1.Container{
		Name:    "runtime",
		Image:   c.runtimeImage,
		Command: c.compileRuntimeCommand(vm),
	}
}

const initContainerName = "executor"

func (c *compiler) compileInitContainer() corev1.Container {
	return corev1.Container{
		Name:  initContainerName,
		Image: c.executorImage,
		Command: []string{
			"/bin/sh", "-c",
			"cp ${EXECUTOR_SOURCE_PATH} ${EXECUTOR_TARGET_PATH} && chmod +x ${EXECUTOR_TARGET_PATH}",
		},
		Args:            nil,
		WorkingDir:      "",
		Ports:           nil,
		EnvFrom:         nil,
		ImagePullPolicy: corev1.PullAlways,
		Env: []corev1.EnvVar{
			{Name: "EXECUTOR_SOURCE_PATH", Value: c.executorPath},
			{Name: "EXECUTOR_TARGET_PATH", Value: executorVolumePath},
		},
		Resources:    corev1.ResourceRequirements{},
		VolumeMounts: c.getSystemVolumeMounts(),
	}
}

func (c *compiler) compileRuntimeCommand(vm *api.VM) []string {
	command := []string{
		"qemu-system-x86_64",
		"-nodefaults",
		"--nographic",
		"-enable-kvm",
		"-display none",
		"-machine type=pc,usb=off",
		fmt.Sprintf("-smp %d,sockets=1,cores=%d,threads=1", vm.CPU, vm.CPU),
		fmt.Sprintf("-m %dM -device virtio-balloon-pci,id=balloon0", vm.Memory),
	}
	return command
}

func (c *compiler) compileVolumes() []corev1.Volume {
	return c.getSystemVolumes()
}

const (
	executorVolumeName = "runner-executor"
	executorVolumePath = "/executor-bin/executor"
)

func (c *compiler) getSystemVolumes() []corev1.Volume {
	return []corev1.Volume{{
		Name: executorVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}}
}

func (c *compiler) getSystemVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{{
		Name:      executorVolumeName,
		ReadOnly:  false,
		MountPath: filepath.Dir(executorVolumePath),
	}}
}
