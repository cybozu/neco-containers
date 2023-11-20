package main

import (
	"os"
	"time"

	"github.com/go-yaml/yaml"
)

const defaultFetchInterval = 30 * time.Second

type Config struct {
	MapNames      []string      `yaml:"mapNames"`
	FetchInterval time.Duration `yaml:"fetchInterval"`
}

func loadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	if cfg.FetchInterval == 0 {
		cfg.FetchInterval = defaultFetchInterval
	}
	return &cfg, nil
}
