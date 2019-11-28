package main

import (
	"net/http"
	"net/url"
	"testing"
)

func TestDirectInner(t *testing.T) {
	cases := []struct {
		name         string
		host         string
		inner        uint16
		resolveMap   map[string]string
		expectedHost string
	}{
		{
			name:         "ResolveIPAddressFromHost",
			host:         "cybozu.com",
			inner:        80,
			resolveMap:   map[string]string{"cybozu": "10.1.2.3"},
			expectedHost: "10.1.2.3:80",
		},
		{
			name:         "ResolveIPAddressFromHostWithEnvironment",
			host:         "stage0-boot-0",
			inner:        443,
			resolveMap:   map[string]string{"boot-0": "10.1.2.3"},
			expectedHost: "10.1.2.3:443",
		},
		{
			name:         "ResolveIPAddressFromHostNotExist",
			host:         "notexist.com",
			inner:        80,
			resolveMap:   map[string]string{"cybozu": "10.1.2.3"},
			expectedHost: "0.0.0.0:80",
		},
		{
			name:         "ResolveIPAddressFromHostWithResolveMapHasNotSet",
			host:         "cybozu.com",
			inner:        80,
			resolveMap:   nil,
			expectedHost: "0.0.0.0:80",
		},
		{
			name:         "ResolveIPAddressFromEmptyHost",
			host:         "",
			inner:        80,
			resolveMap:   map[string]string{"cybozu": "10.1.2.3"},
			expectedHost: "0.0.0.0:80",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			request := &http.Request{Host: c.host, URL: &url.URL{}}
			directorToInner(request, c.inner, c.resolveMap)

			if request.URL.Host != c.expectedHost {
				t.Errorf("request.URL.Host != c.expectedHost %s %s", request.URL.Host, c.expectedHost)
			}

			if request.URL.Scheme != "https" {
				t.Errorf("request.URL.Scheme != \"https\" %s", request.URL.Scheme)
			}
		})
	}
}
