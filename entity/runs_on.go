package entity

type RunsOn struct {
	Label  string  `json:"label"`
	Docker *Docker `json:"docker"`
}

type Docker struct {
	Containers       []Container `json:"containers"`
	Volumes          []Volume    `json:"volumes"`
	DefaultContainer string      `json:"default_container"`
}

type Container struct {
	Name         string         `json:"name"`
	Image        string         `json:"image"`
	VolumeMounts []*VolumeMount `json:"volume_mounts"`
}

type Volume struct {
	Name     string                `json:"name"`
	HostPath *HostPathVolumeSource `json:"host_path"`
	EmptyDir *EmptyDirVolumeSource `json:"empty_dir"`
}

type HostPathVolumeSource struct {
	Path string `json:"path"`
}

type EmptyDirVolumeSource struct{}

type VolumeMount struct {
	Name      string `json:"name"`
	ReadOnly  bool   `json:"read_only"`
	MountPath string `json:"mount_path"`
}
