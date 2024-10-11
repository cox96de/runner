package vm

import (
	"testing"

	"gotest.tools/v3/assert"
	corev1 "k8s.io/api/core/v1"
)

func Test_parseVolumes(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		volumes, mounts, err := parseVolumes([]string{"type=hostPath,name=mnt,mountPath=/mnt,path=/root/mnt", "type=nfs,name=nfs,mountPath=/mnt,server=192.168.33.1,path=/nfs"})
		assert.NilError(t, err)
		assert.DeepEqual(t, volumes, []corev1.Volume{
			{
				Name: "mnt",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/root/mnt",
					},
				},
			},
			{
				Name: "nfs",
				VolumeSource: corev1.VolumeSource{
					NFS: &corev1.NFSVolumeSource{
						Path:   "/nfs",
						Server: "192.168.33.1",
					},
				},
			},
		})
		assert.DeepEqual(t, mounts, []corev1.VolumeMount{
			{
				Name:      "mnt",
				MountPath: "/mnt",
			},
			{
				Name:      "nfs",
				MountPath: "/mnt",
			},
		})
	})
	t.Run("bad", func(t *testing.T) {
		_, _, err := parseVolumes([]string{"type=hostPath,mountPath=/mnt,path=/root/mnt"})
		assert.ErrorContains(t, err, "name is required")
	})
	t.Run("unknown_type", func(t *testing.T) {
		_, _, err := parseVolumes([]string{"type=unknown,mountPath=/mnt,path=/root/mnt"})
		assert.ErrorContains(t, err, "unsupported volume type")
	})
}
