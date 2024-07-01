package client

import (
	"fmt"
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
	if c.RetryNum < 0 {
		return fmt.Errorf("retry must be greater than or equal to 0 (input: %d)", c.RetryNum)
	}
	return nil
}
