package main

import (
	"github.com/cox96de/containervm/cloudinit"
	"gopkg.in/yaml.v3"
)

type CloudInitV1 struct {
	Version int       `yaml:"version"`
	Config  []*Config `yaml:"config"`
}

type Config struct {
	Type    string    `yaml:"type"`
	Name    string    `yaml:"name"`
	MacAddr string    `yaml:"mac_address"`
	Subnets []*Subnet `yaml:"subnets"`
}

type Subnet struct {
	Type    string `yaml:"type"`
	Address string `yaml:"address"`
	Gateway string `yaml:"gateway"`
}

func GenerateNetworkConfig(config *cloudinit.NetworkConfig) ([]byte, error) {
	cc := &Config{
		Type:    "physical",
		Name:    "net0",
		MacAddr: config.Mac.String(),
	}
	if config.Gateway4 != nil {
		var addr string
		for _, address := range config.Addresses {
			if address.IP.To4() != nil {
				addr = address.String()
			}
		}
		cc.Subnets = append(cc.Subnets, &Subnet{
			Type:    "static",
			Address: addr,
			Gateway: config.Gateway4.String(),
		})
	}
	if config.Gateway6 != nil {
		var addr string
		for _, address := range config.Addresses {
			if address.IP.IsGlobalUnicast() {
				addr = address.String()
			}
		}
		cc.Subnets = append(cc.Subnets, &Subnet{
			Type:    "static",
			Address: addr,
			Gateway: config.Gateway6.String(),
		})
	}
	// Windows cloudbase-init only support v1 config.
	c := &CloudInitV1{
		Version: 1,
		Config: []*Config{
			cc,
		},
	}
	out, err := yaml.Marshal(c)
	if err != nil {
		return nil, err
	}
	return append([]byte("#cloud-config\n"), out...), nil
}
