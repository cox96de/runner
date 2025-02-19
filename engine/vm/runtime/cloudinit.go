package main

import (
	"github.com/cox96de/containervm/cloudinit"
	"github.com/cox96de/runner/util"
	"gopkg.in/yaml.v3"
)

type CloudInitV1 struct {
	Version int       `yaml:"version"`
	Config  []*Config `yaml:"config"`
}

type Config struct {
	Type    string    `yaml:"type,omitempty"`
	Name    string    `yaml:"name,omitempty"`
	MacAddr string    `yaml:"mac_address,omitempty"`
	Subnets []*Subnet `yaml:"subnets,omitempty"`
	Address []string  `yaml:"address,omitempty"`
	Search  []string  `yaml:"search,omitempty"`
}

type Subnet struct {
	Type    string `yaml:"type"`
	Address string `yaml:"address"`
	Gateway string `yaml:"gateway"`
}

type NetworkConfig struct {
	cloudinit.NetworkConfig
	Nameservers   []string
	SearchDomains []string
}

func GenerateNetworkConfig(config *NetworkConfig) ([]byte, error) {
	cc := &Config{
		Type: "physical",
		// In windows, occurs
		//   cloudbase init '{Object Exists} An attempt was made to create an object and the object name already existed.
		//   ' 'Renaming interface \"Ethernet\" to \"net0\" failed': cloudbaseinit.exception.CloudbaseInitException: Renaming interface \"Ethernet\" to \"net0\" failed
		// Create a random interface name.
		Name:    "net" + util.RandomLower(2),
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
	if len(config.Nameservers) > 0 {
		c.Config = append(c.Config, &Config{
			Type:    "nameserver",
			Address: config.Nameservers,
			Search:  config.SearchDomains,
		})
	}
	out, err := yaml.Marshal(c)
	if err != nil {
		return nil, err
	}
	return append([]byte("#cloud-config\n"), out...), nil
}
