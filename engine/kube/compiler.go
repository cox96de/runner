package kube

import (
	"path/filepath"
	"strconv"
	"sync/atomic"

	"github.com/cox96de/runner/api"

	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type compiler struct {
	executorImage   string
	executorPath    string
	executorPortMap map[string]int32
	portCount       atomic.Int32
}

func newCompiler(executorImage string, executorPath string) *compiler {
	c := &compiler{
		executorImage: executorImage, executorPath: executorPath, executorPortMap: map[string]int32{},
		portCount: atomic.Int32{},
	}
	c.portCount.Store(1234)
	return c
}

type compileResult struct {
	pod *corev1.Pod
}

func (c *compiler) Compile(id string, spec *api.RunsOn) *compileResult {
	result := &compileResult{
		pod: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: id,
			},
			Spec: corev1.PodSpec{
				InitContainers: []corev1.Container{c.compileInitContainer()},
				Containers:     c.compileContainers(spec.Docker.Containers),
				Volumes:        c.compileVolumes(spec.Docker.Volumes),
			},
		},
	}
	return result
}

func (c *compiler) compileContainers(containers []*api.Container) []corev1.Container {
	return lo.Map(containers, func(container *api.Container, index int) corev1.Container {
		return c.compileContainer(container)
	})
}

func (c *compiler) compileContainer(container *api.Container) corev1.Container {
	return corev1.Container{
		Name:         container.Name,
		Image:        container.Image,
		Command:      c.compileContainerCommand(container.Name),
		VolumeMounts: c.compileContainerVolumeMounts(container.VolumeMounts),
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

func (c *compiler) compileContainerCommand(containerID string) []string {
	return []string{
		executorVolumePath,
		"--port", c.getNextPort(containerID),
	}
}

func (c *compiler) getNextPort(containerID string) string {
	port := c.portCount.Add(1)
	c.executorPortMap[containerID] = port
	return strconv.FormatInt(int64(port), 10)
}

func (c *compiler) compileContainerVolumeMounts(volumeMounts []*api.VolumeMount) []corev1.VolumeMount {
	return append(lo.Map(volumeMounts, func(item *api.VolumeMount, index int) corev1.VolumeMount {
		return c.compileVolumeMount(item)
	}), c.getSystemVolumeMounts()...)
}

func (c *compiler) compileVolumeMount(v *api.VolumeMount) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      v.Name,
		ReadOnly:  v.ReadOnly,
		MountPath: v.MountPath,
	}
}

func (c *compiler) compileVolumes(volumes []*api.Volume) []corev1.Volume {
	return append(lo.Map(volumes, func(item *api.Volume, index int) corev1.Volume {
		return c.compileVolume(item)
	}), c.getSystemVolumes()...)
}

func (c *compiler) compileVolume(v *api.Volume) corev1.Volume {
	volume := corev1.Volume{
		Name:         v.Name,
		VolumeSource: corev1.VolumeSource{},
	}
	if v.EmptyDir != nil {
		volume.EmptyDir = &corev1.EmptyDirVolumeSource{}
	}
	if v.HostPath != nil {
		volume.HostPath = &corev1.HostPathVolumeSource{}
	}
	return volume
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
