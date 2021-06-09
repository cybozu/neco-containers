package main

import (
	"net"
	"strings"
	"testing"
)

func TestParseCIDRs(t *testing.T) {
	nets, err := parseCIDRs("", "hoge")
	if err != nil || len(nets) != 0 {
		t.Errorf("nets: %v, err: %v", nets, err)
	}

	nets, err = parseCIDRs(", ,", "hoge")
	if err != nil || len(nets) != 0 {
		t.Errorf("nets: %v, err: %v", nets, err)
	}

	nets, err = parseCIDRs("1.2.3.4/30, 9.8.7.64/26", "hoge")
	if err != nil || len(nets) != 2 || nets[0].String() != "1.2.3.4/30" || nets[1].String() != "9.8.7.64/26" {
		t.Errorf("nets: %v, err: %v", nets, err)
	}

	nets, err = parseCIDRs("0.0.0.0/33", "hoge")
	if err == nil || !strings.Contains(err.Error(), " hoge ") {
		t.Errorf("nets: %v, err: %v", nets, err)
	}
}

func TestParseAllowDeny(t *testing.T) {
	ad, err := ParseAllowDeny("1.2.3.0/24", "1.2.0.0/16,5:6::/32")
	if err != nil {
		t.Fatalf("ad: %v, err: %v", ad, err)
	}

	for _, a := range []string{"1.2.3.4", "5.6.7.8", "1:2:3:4::"} {
		if !ad.IsAllowedIP(net.ParseIP(a)) {
			t.Errorf("%s should be allowed", a)
		}
		if !ad.IsAllowedAddr(a) {
			t.Errorf("%s should be allowed", a)
		}
	}
	for _, a := range []string{"1.2.2.1", "5:6:7::", "hoge"} {
		if ad.IsAllowedIP(net.ParseIP(a)) {
			t.Errorf("%s should not be allowed", a)
		}
		if ad.IsAllowedAddr(a) {
			t.Errorf("%s should not be allowed", a)
		}
	}
	if ad.IsAllowedIP(nil) {
		t.Errorf("nil should not be allowed")
	}

	if !ad.IsAllowedHostPort("1.2.3.4:9876") {
		t.Errorf("1.2.3.4:9876 should be allowed")
	}
}
