package vm

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cox96de/runner/api"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	executorMountTag   = "_executor"
	executorMountPoint = "/executor"
	executorPort       = 1234
)

type compiler struct {
	runtimeImage        string
	executorPath        string
	imageBaseDir        string
	runtimeVolumeMounts []corev1.VolumeMount
	runtimeVolumes      []corev1.Volume
}

func newCompiler(runtimeImage string, executorPath string, imageBaseDir string, runtimeVolumeMounts []corev1.VolumeMount,
	runtimeVolumes []corev1.Volume,
) *compiler {
	return &compiler{
		runtimeImage: runtimeImage, executorPath: executorPath, imageBaseDir: imageBaseDir,
		runtimeVolumeMounts: runtimeVolumeMounts, runtimeVolumes: runtimeVolumes,
	}
}

type compileResult struct {
	pod *corev1.Pod
}

func (c *compiler) Compile(id string, spec *api.RunsOn) (*compileResult, error) {
	containers, err := c.compileContainers(spec.VM.Image)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to compile container")
	}
	result := &compileResult{
		pod: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: id,
			},
			Spec: corev1.PodSpec{
				Containers: containers,
				Volumes:    c.compileVolumes(),
			},
		},
	}
	return result, nil
}

func (c *compiler) compileContainers(imageName string) ([]corev1.Container, error) {
	q := newQemuCompiler(2, 1024)
	q.AddDisk(filepath.Join(c.imageBaseDir, imageName))
	q.AddShare(filepath.Dir(c.executorPath), executorMountTag)
	metaData := `instance-id: vm-runner
local-hostname: vm-runner
`
	executorPathInVM := filepath.Join(executorMountPoint, filepath.Base(c.executorPath))
	data := cloudInitUserData{
		RunCMD: [][]string{
			{
				"sh",
				"-c",
				fmt.Sprintf("while true; do if [ -f %s ]; then nohup %s --port %d > /var/executor.log 2>&1 & break; else echo \"Executor binary file not found. Retrying in 1 second...\"; sleep 1; fi; done",
					executorPathInVM, executorPathInVM, executorPort),
			},
		},
		Mounts: [][]string{
			{
				executorMountTag, executorMountPoint, "9p", "trans=virtio,version=9p2000.L,msize=104857600", "0", "0",
			},
		},
	}
	userData, err := data.Marshal()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to generate cloud init user data")
	}
	container := corev1.Container{
		Name:            "vm-runtime",
		Image:           c.runtimeImage,
		ImagePullPolicy: corev1.PullAlways,
		Args:            append([]string{"--"}, q.Compile()...),
		Env: []corev1.EnvVar{
			{
				Name:  "CLOUD_INIT_USER_DATA",
				Value: userData,
			},
			{
				Name:  "CLOUD_INIT_META_DATA",
				Value: metaData,
			},
		},
		SecurityContext: &corev1.SecurityContext{
			Privileged: lo.ToPtr(true),
		},
		VolumeMounts: c.compileContainerVolumeMounts(),
	}
	return []corev1.Container{container}, nil
}

func (c *compiler) compileContainerVolumeMounts() []corev1.VolumeMount {
	return c.getSystemVolumeMounts()
}

func (c *compiler) getSystemVolumeMounts() []corev1.VolumeMount {
	return c.runtimeVolumeMounts
}

func (c *compiler) compileVolumes() []corev1.Volume {
	return c.getSystemVolumes()
}

func (c *compiler) getSystemVolumes() []corev1.Volume {
	return c.runtimeVolumes
}

// qemuCompiler is to compile qemu command
type qemuCompiler struct {
	cpu       int
	memory    int // MB
	disks     []*disk
	shares    []*shareVolume
	enableVGA bool
}

func newQemuCompiler(cpu int, memory int) *qemuCompiler {
	return &qemuCompiler{cpu: cpu, memory: memory}
}

type disk struct {
	path string
}

type shareVolume struct {
	path string
	tag  string
}

func (c *qemuCompiler) AddDisk(path string) {
	c.disks = append(c.disks, &disk{path})
}

func (c *qemuCompiler) AddShare(path string, tag string) {
	c.shares = append(c.shares, &shareVolume{path: path, tag: tag})
}

func (c *qemuCompiler) Compile() []string {
	var cmds []string
	// HARDCODE x86_64
	cmds = append(cmds, "qemu-system-x86_64")
	cmds = append(cmds, "-nodefaults")
	cmds = append(cmds, "--nographic")
	cmds = append(cmds, "-display", "none")
	cmds = append(cmds, "-machine", "type=q35,usb=off")
	cmds = append(cmds, "--enable-kvm")
	cmds = append(cmds, "-cpu", "host")

	cmds = append(cmds, "-smp", fmt.Sprintf("%d,sockets=1,cores=%d,threads=1", c.cpu, c.cpu))
	cmds = append(cmds, "-m", fmt.Sprintf("%dM", c.memory), "-device", "virtio-balloon-pci,id=balloon0")
	for _, disk := range c.disks {
		cmds = append(cmds, "-drive", fmt.Sprintf("file=%s,format=qcow2,if=virtio,aio=threads,media=disk,cache=unsafe,snapshot=on", disk.path))
	}
	for idx, share := range c.shares {
		cmds = append(cmds, "-fsdev", fmt.Sprintf("local,security_model=passthrough,id=fsdev%d,path=%s", idx, share.path))
		cmds = append(cmds, "-device", fmt.Sprintf("virtio-9p-pci,fsdev=fsdev%d,mount_tag=%s", idx, share.tag))
	}
	cmds = append(cmds, "-serial", "chardev:serial0", "-chardev", "socket,id=serial0,server=on,wait=off,path=/tmp/console.sock")
	if c.enableVGA {
		cmds = append(cmds, "-device", "VGA")
	}
	return cmds
}

func parseVolumes(exprs []string) ([]corev1.Volume, []corev1.VolumeMount, error) {
	var volumes []corev1.Volume
	var mounts []corev1.VolumeMount
	for _, expr := range exprs {
		volume, volumeMount, err := parseVolume(expr)
		if err != nil {
			return nil, nil, err
		}
		volumes = append(volumes, volume)
		mounts = append(mounts, volumeMount)
	}
	return volumes, mounts, nil
}

func parseVolume(expr string) (_ corev1.Volume, _ corev1.VolumeMount, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Errorf("%+v", r)
			return
		}
	}()
	split := strings.Split(expr, ",")
	kv := map[string]string{}
	for _, s := range split {
		n := strings.SplitN(s, "=", 2)
		if len(n) == 1 {
			kv[n[0]] = ""
			continue
		}
		if len(n) < 2 {
			return corev1.Volume{}, corev1.VolumeMount{}, fmt.Errorf("invalid volume express %s", expr)
		}
		kv[n[0]] = n[1]
	}
	volumeType := expectString(kv, "type", "volume type is required")
	switch volumeType {
	case "hostPath":
		volume := corev1.Volume{
			Name: expectString(kv, "name", "name is required for hostPath volume"),
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: expectString(kv, "path", "path is required for hostPath volume"),
				},
			},
		}
		volumeMount := corev1.VolumeMount{
			Name:      expectString(kv, "name", "name is required for hostPath volume"),
			MountPath: expectString(kv, "mountPath", "mountPath is required for hostPath volume"),
		}
		return volume, volumeMount, nil
	case "nfs":
		volume := corev1.Volume{
			Name: kv["name"],
			VolumeSource: corev1.VolumeSource{
				NFS: &corev1.NFSVolumeSource{
					Server:   expectString(kv, "server", "server is required for nfs volume"),
					Path:     expectString(kv, "path", "path is required for nfs volume"),
					ReadOnly: getBool(kv, "readOnly", false),
				},
			},
		}
		volumeMount := corev1.VolumeMount{
			Name:      expectString(kv, "name", "name is required for nfs volume"),
			MountPath: expectString(kv, "mountPath", "mountPath is required for nfs volume"),
		}
		return volume, volumeMount, nil
	}
	return corev1.Volume{}, corev1.VolumeMount{}, errors.Errorf("unsupported volume type %s", volumeType)
}

func getBool(kv map[string]string, key string, defaultValue bool) bool {
	v, ok := kv[key]
	if !ok {
		return defaultValue
	}
	switch v {
	case "true":
		return true
	case "false":
		return false
	case "":
		return true
	default:
		panic(fmt.Sprintf("expect boolean value for %s, got %s", key, v))
	}
}

func expectString(kv map[string]string, key string, message string) string {
	v, ok := kv[key]
	if !ok {
		panic(message)
	}
	return v
}
