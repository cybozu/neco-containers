package config

import (
	"net/netip"

	"sigs.k8s.io/yaml"
)

type SourceIPConfig struct {
	AllowedCIDRs []string `json:"allowedCIDRs,omitempty"`
}

type Config struct {
	SourceIP SourceIPConfig `json:"sourceIP,omitempty"`
}

func (c *Config) Load(data []byte) error {
	err := yaml.Unmarshal(data, c, yaml.DisallowUnknownFields)
	if err != nil {
		return err
	}
	c.FillDefaults()
	return c.Validate()
}

func (c *Config) FillDefaults() {
}

func (c *Config) Validate() error {
	if c.SourceIP.AllowedCIDRs != nil {
		for _, cidr := range c.SourceIP.AllowedCIDRs {
			_, err := netip.ParsePrefix(cidr)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
