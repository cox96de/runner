package vm

import "gopkg.in/yaml.v3"

// cloudInitUserData only contains neccesary plugins we need.
type cloudInitUserData struct {
	RunCMD [][]string `yaml:"runcmd"`
	Mounts [][]string `yaml:"mounts"`
}

func (c *cloudInitUserData) Marshal() (string, error) {
	output, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	return "#cloud-config\n" + string(output), nil
}
