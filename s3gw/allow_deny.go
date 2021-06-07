package main

import (
	"fmt"
	"net"
	"strings"
)

// AllowDeny represents address based access control like hosts.allow
type AllowDeny struct {
	allowed []*net.IPNet
	denied  []*net.IPNet
}

func parseCIDRs(cidrs, name string) ([]*net.IPNet, error) {
	var nets []*net.IPNet

	for _, s := range strings.Split(cidrs, ",") {
		s = strings.Trim(s, " ")
		if s == "" {
			continue
		}
		_, n, err := net.ParseCIDR(s)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s CIDR: %w", name, err)
		}
		nets = append(nets, n)
	}

	return nets, nil
}

// ParseAllowDeny parses ACL. Note: it only accepts CIDR. (hostnames are not supported)
func ParseAllowDeny(allowed, denied string) (*AllowDeny, error) {
	var ad AllowDeny

	nets, err := parseCIDRs(allowed, "allowed")
	if err != nil {
		return nil, err
	}
	ad.allowed = nets

	nets, err = parseCIDRs(denied, "denied")
	if err != nil {
		return nil, err
	}
	ad.denied = nets

	return &ad, nil
}

func (ad *AllowDeny) IsAllowedHostPort(hostport string) bool {
	addr, _, err := net.SplitHostPort(hostport)
	if err != nil {
		return false
	}
	return ad.IsAllowedAddr(addr)
}

func (ad *AllowDeny) IsAllowedAddr(addr string) bool {
	ip := net.ParseIP(addr)
	if ip == nil {
		return false
	}
	return ad.IsAllowedIP(ip)
}

func (ad *AllowDeny) IsAllowedIP(ip net.IP) bool {
	if ip == nil {
		return false
	}

	for _, n := range ad.allowed {
		if n.Contains(ip) {
			return true
		}
	}
	for _, n := range ad.denied {
		if n.Contains(ip) {
			return false
		}
	}
	return true
}
