package main

import (
	"fmt"
	"io"
	"net/http"
)

type SquidClient interface {
	GetCounters() (io.ReadCloser, error)
	GetServiceTimes() (io.ReadCloser, error)
}

type squidClient struct {
	client *http.Client
	Host   string
	Port   int
}

func NewSquidClient(config *Config) *squidClient {
	return &squidClient{
		client: &http.Client{},
		Host:   config.SquidHost,
		Port:   config.SquidPort,
	}
}

func (c *squidClient) GetCounters() (io.ReadCloser, error) {
	resp, err := c.client.Get(fmt.Sprintf("http://%s:%d/squid-internal-mgr/counters", c.Host, c.Port))
	if err != nil {
		return nil, err
	}
	return resp.Body, err
}

func (c *squidClient) GetServiceTimes() (io.ReadCloser, error) {
	resp, err := c.client.Get(fmt.Sprintf("http://%s:%d/squid-internal-mgr/service_times", c.Host, c.Port))
	if err != nil {
		return nil, err
	}
	return resp.Body, err
}
