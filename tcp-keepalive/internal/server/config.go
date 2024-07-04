package server

import (
	"net/netip"
)

type Config struct {
	ListenAddr string
}

func (c *Config) Validate() error {
	if _, err := netip.ParseAddrPort(c.ListenAddr); err != nil {
		return err
	}
	return nil
}
