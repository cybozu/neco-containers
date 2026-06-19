package metrics

import "net/netip"

type Config struct {
	Export   bool
	AddrPort string
}

func (c *Config) Validate() error {
	if _, err := netip.ParseAddrPort(c.AddrPort); err != nil {
		return err
	}
	return nil
}
