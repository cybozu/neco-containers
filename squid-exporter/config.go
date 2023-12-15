package main

import "flag"

type Config struct {
	SquidHost   string
	SquidPort   int
	MetricsPort int
}

var (
	flagSquidHost   = flag.String("squid-host", "localhost", "Squid host")
	flagSquidPort   = flag.Int("squid-port", 3128, "Squid port")
	flagMetricsPort = flag.Int("metrics-port", 8080, "Metrics port")
)

func NewConfig() *Config {
	flag.Parse()
	return &Config{
		SquidHost:   *flagSquidHost,
		SquidPort:   *flagSquidPort,
		MetricsPort: *flagMetricsPort,
	}
}
