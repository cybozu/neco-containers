package server

import (
	"net"
)

type Config struct {
	ListenAddr string
}

func (c *Config) Validate() error {
	if _, err := net.ResolveTCPAddr("tcp", c.ListenAddr); err != nil {
		return err
	}
	return nil
}
