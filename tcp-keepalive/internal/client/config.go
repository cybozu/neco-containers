package client

import (
	"net/netip"
	"time"
)

type Config struct {
	ReceiveTimeout time.Duration
	RetryInterval  time.Duration
	RetryNum       int
	SendInterval   time.Duration
	ServerAddr     string
}

func (c *Config) Validate() error {
	if _, err := netip.ParseAddrPort(c.ServerAddr); err != nil {
		return err
	}
	return nil
}
