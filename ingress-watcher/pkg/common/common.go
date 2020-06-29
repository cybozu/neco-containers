package common

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	flag "github.com/spf13/pflag"
)

// WatchConfig is a configuration used by export and push subcommands.
type WatchConfig struct {
	TargetURLs     []string
	WatchInterval  time.Duration
	PermitInsecure bool
	ResolveRules   []string
}

// SetCommonFlags sets common flags decoder using cobra.
func (c *WatchConfig) SetCommonFlags(fs *flag.FlagSet) {
	fs.StringArrayVarP(&c.TargetURLs, "target-urls", "", nil, "Target Ingress address and port.")
	fs.DurationVarP(&c.WatchInterval, "watch-interval", "", 5*time.Second, "Watching interval.")
	fs.BoolVar(&c.PermitInsecure, "permit-insecure", false, "Permit insecure access to targets.")
	fs.StringArrayVarP(&c.ResolveRules, "resolve-rules", "", nil, "Resolve rules from FQDN to IPv4 address (ex. example.com:192.168.0.1).")
}

// CheckCommonFlags checks common flags.
func (c *WatchConfig) CheckCommonFlags() error {
	for _, rule := range c.ResolveRules {
		split := strings.Split(rule, ":")
		if len(split) != 2 {
			return errors.New(`invalid format in "resolve-rules" : ` + rule)
		}
	}

	if len(c.TargetURLs) == 0 {
		return errors.New(`required flag "target-urls" not set`)
	}

	return nil
}

// GetClient generates http.Client from the configuration.
func (c *WatchConfig) GetClient() *http.Client {
	var transport *http.Transport
	if c.PermitInsecure {
		if transport == nil {
			transport = &http.Transport{}
		}
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if len(c.ResolveRules) > 0 {
		resolveMap := make(map[string]string)
		for _, rules := range c.ResolveRules {
			s := strings.Split(rules, ":")
			resolveMap[s[0]] = s[1]
		}

		dialerFunc := func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			splitAddr := strings.Split(address, ":")
			if len(splitAddr) > 2 {
				return nil, errors.New(`invalid format : ` + address)
			}

			if ip, ok := resolveMap[splitAddr[0]]; ok {
				return d.DialContext(ctx, network, ip+":"+splitAddr[1])
			}
			return d.DialContext(ctx, network, address)
		}

		if transport == nil {
			transport = &http.Transport{}
		}
		transport.DialContext = dialerFunc
	}

	client := &http.Client{}
	if transport != nil {
		client.Transport = transport
	}

	return client
}
