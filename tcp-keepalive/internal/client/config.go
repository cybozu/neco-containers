package client

import (
	"net"
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
	if _, err := net.ResolveTCPAddr("tcp", c.ServerAddr); err != nil {
		return err
	}
	return nil
}
